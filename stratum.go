package stratum

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/fatih/set"
	log "github.com/sirupsen/logrus"
)

var (
	KeepAliveDuration time.Duration = 5 * time.Second
)

type StratumOnWorkHandler func(work *Work)
type StratumContext struct {
	net.Conn
	reader            *bufio.Reader
	id                int
	SessionID         string
	KeepAliveDuration time.Duration
	Work              *Work
	workListeners     set.Interface
	submitListeners   set.Interface
	responseListeners set.Interface
	LastSubmittedWork *Work
}

func New() *StratumContext {
	sc := &StratumContext{}
	sc.KeepAliveDuration = KeepAliveDuration
	sc.workListeners = set.New()
	sc.submitListeners = set.New()
	sc.responseListeners = set.New()
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
	line, err := sc.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func (sc *StratumContext) ReadJSON() (map[string]interface{}, error) {
	line, err := sc.reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	var ret map[string]interface{}
	if err = json.Unmarshal([]byte(line), &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (sc *StratumContext) ReadResponse() (*Response, error) {
	line, err := sc.ReadLine()
	if err != nil {
		return nil, err
	}
	line = strings.TrimSpace(line)
	log.Debugf("Server sent back: %v", line)
	return ParseResponse([]byte(line))
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
		if work, err := ParseWork(response.Result["job"].(map[string]interface{})); err != nil {
			return err
		} else {
			log.Infof("Stratum detected new block")
			sc.NotifyNewWork(work)
		}
	}

	go func() {
		for {
			line, err := sc.ReadLine()
			if err != nil {
				log.Errorf("Failed to read string from stratum: %v", err)
				continue
			}
			log.Debugf("Received line from server: %v", line)

			var msg map[string]interface{}
			if err = json.Unmarshal([]byte(line), &msg); err != nil {
				log.Errorf("Failed to unmarshal line into JSON: %v", err)
				continue
			}

			if _, ok := msg["id"]; ok {
				// This is a response
				response, err := ParseResponse([]byte(line))
				if err != nil {
					log.Errorf("Failed to parse response from server: %v", err)
				} else {
					sc.NotifyResponse(response)
				}
			} else {
				log.Debugf("Received message from stratum server: %v", msg)
				switch msg["method"].(string) {
				case "job":
					if work, err := ParseWork(msg["params"].(map[string]interface{})); err != nil {
						log.Errorf("Failed to parse job: %v", err)
						continue
					} else {
						sc.NotifyNewWork(work)
					}
				default:
				}
			}
		}
	}()

	// Keep-alive
	go func() {
		for {
			time.Sleep(sc.KeepAliveDuration)
			args := make(map[string]interface{})
			args["id"] = sc.SessionID
			if err := sc.Call("keepalive", args); err != nil {
				log.Errorf("Failed keepalive: %v", err)
			} else {
				// log.Debugf("Posted keepalive")
			}
		}
	}()

	return nil
}

func (sc *StratumContext) SubmitWork(work *Work, hash string) error {
	if work == sc.LastSubmittedWork {
		log.Warnf("Prevented submission of stale work")
		return nil
	}
	args := make(map[string]interface{})
	nonceStr, err := BinToHex(work.Data[39:43])
	if err != nil {
		return err
	}
	args["id"] = sc.SessionID
	args["job_id"] = work.JobID
	args["nonce"] = nonceStr
	args["result"] = hash
	if err := sc.Call("submit", args); err != nil {
		return err
	} else {
		// Successfully submitted result
		log.Debugf("Successfully submitted work result")
		args["work"] = work
		sc.NotifySubmit(args)
		sc.LastSubmittedWork = work
	}
	return nil
}

func (sc *StratumContext) RegisterSubmitListener(sChan chan interface{}) {
	log.Debugf("Registerd stratum.submitListener")
	sc.submitListeners.Add(sChan)
}

func (sc *StratumContext) RegisterWorkListener(workChan chan *Work) {
	log.Debugf("Registerd stratum.workListener")
	sc.workListeners.Add(workChan)
}

func (sc *StratumContext) RegisterResponseListener(rChan chan *Response) {
	log.Debugf("Registerd stratum.responseListener")
	sc.responseListeners.Add(rChan)
}

func (sc *StratumContext) GetJob() error {
	args := make(map[string]interface{})
	args["id"] = sc.SessionID
	return sc.Call("getjob", args)
}

func ParseResponse(b []byte) (*Response, error) {
	var response Response
	if err := json.Unmarshal(b, &response); err != nil {
		return nil, err
	}
	return &response, nil
}

func (sc *StratumContext) NotifyNewWork(work *Work) {
	sc.Work = work
	for _, obj := range sc.workListeners.List() {
		ch := obj.(chan *Work)
		ch <- work
	}
}

func (sc *StratumContext) NotifySubmit(data interface{}) {
	for _, obj := range sc.submitListeners.List() {
		ch := obj.(chan interface{})
		ch <- data
	}
}

func (sc *StratumContext) NotifyResponse(response *Response) {
	for _, obj := range sc.responseListeners.List() {
		ch := obj.(chan *Response)
		ch <- response
	}
}
