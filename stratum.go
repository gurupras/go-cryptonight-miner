package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

type StratumContext struct {
	net.Conn
	reader       *bufio.Reader
	id           int
	ResponseChan chan *Response
	SessionID    string
	WorkChan     chan *Work
}

func New() *StratumContext {
	sc := &StratumContext{}
	sc.ResponseChan = make(chan *Response, 0)
	sc.WorkChan = make(chan *Work, 0)
	return sc
}

func (sc *StratumContext) Connect(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	log.Debugf("Dial success")
	sc.Conn = conn
	sc.reader = bufio.NewReader(conn)
	return nil
}

func (sc *StratumContext) Subscribe() error {
	if err := sc.Call(RPC_SUBSCRIBE_METHOD, []string{}); err != nil {
		return fmt.Errorf("Failed to subscribe: %v", err)
	}
	return nil
}

func (sc *StratumContext) Call(serviceMethod string, args interface{}) error {
	sc.id++

	req := NewRequest(sc.id, serviceMethod, args)
	str, err := req.JsonRPCString()
	if err != nil {
		return err
	}
	if _, err := sc.Write([]byte(str)); err != nil {
		return err
	}
	return nil
}

func (sc *StratumContext) ReadLine() (string, error) {
	return sc.reader.ReadString('\n')
}

func (sc *StratumContext) ReadResponse() (*Response, error) {
	line, err := sc.ReadLine()
	if err != nil {
		return nil, err
	}
	var response Response
	if err := json.Unmarshal([]byte(strings.TrimSpace(line)), &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (sc *StratumContext) Authorize(username, password string) error {
	log.Debugf("Beginning authorize")
	args := make(map[string]interface{})
	args["login"] = username
	args["pass"] = password
	args["agent"] = "go-stratum-client"

	err := sc.Call("login", args)
	if err != nil {
		return err
	}

	response, err := sc.ReadResponse()
	if err != nil {
		return err
	}
	if response.Error != nil {
		return response.Error
	} else {
		log.Infof("Authorization successful")
		sc.SessionID = response.Result["id"].(string)
		// TODO: This also contains a job? We may have to handle it
	}

	if work, err := ParseWorkFromResponse(response); err != nil {
		return err
	} else {
		sc.WorkChan <- work
	}

	go func() {
		for {
			response, err := sc.ReadResponse()
			sc.ResponseChan <- response
			if err != nil {
				log.Errorf("Failed to read string from stratum: %v", err)
			} else {
				log.Debugf("Received message from stratum server: %v", response)
				if work, err := ParseWorkFromResponse(response); work != nil && err == nil {
					sc.WorkChan <- work
				}
			}
		}
	}()
	return nil
}
