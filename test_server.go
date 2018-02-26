package stratum

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"

	stoppablenetlistener "github.com/gurupras/go-stoppable-net-listener"
	log "github.com/sirupsen/logrus"
)

type TestServer struct {
	*stoppablenetlistener.StoppableNetListener
	RequestChan chan *ClientRequest
}

type ClientRequest struct {
	Conn    net.Conn
	Request *Request
}

var (
	TEST_JOB_STR_1   = `{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335","job":{"blob":"0101c1ef8dd405b9a2d3278fbc35ba876422c82a94dd7befec695f42848d55cab06a96e31a34b300000000b1940501101bd2f22938ce55f25ddf3c584ab6915453073f3c09a5967df6966204","job_id":"Js5ps3OcKxJUCiEjtIz54ImuNMmA","target":"b88d0600","id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335"},"status":"OK"}}` + "\n"
	TEST_JOB_STR_2   = `{"jsonrpc":"2.0","method":"job","params":{"blob":"0101dab597d4059fdcc43a65bca7d58238708e97dbe59f21030314c55278c42cbb9ae13ac2e44b00000000b0d68bd268662790c0aae0e79bbdd6c4fd6dabf11485415239930936708e38df07","job_id":"jhUv6SY9RB0Pv+QzyfoZ9sg0Yg1d","target":"877d0200","id":"7ec63ee3-21ae-45ee-abd7-fc44c01508e7"}}` + "\n"
	TEST_JOB_STR_3   = `{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335","job":{"blob":"0505efcfdccb0506180897d587b02f9c97037e66ea638990b2b3a0efab7bab0bff4e3f3dfe1c7d00000000a6788e66eb9b82325f95fc7a2007d3fed7152a3590366cc2a9577dcadf3544a804","job_id":"Js5ps3OcKxJUCiEjtIz54ImuNMmA","target":"e4a63d00","id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335"},"status":"OK"}}` + "\n"
	RESULT_JOB_STR_3 = "960A7A3A1826B0AA70E8043FFE7B9E23EE2E028BBA75F3D7557CCDFF9C7F1A00"

	TEST_JOB_STR_4   = `{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335","job":{"blob":"0707dde5cdd4058e415b279f8e448abc8cf5c97e5768770ad1f2699f932be50511a0925afa9df5000000007d39a83b02d50b39b6ee526555b9bef8d249e03e5871cf3e7a07fba548db00f303","job_id":"K8itLau43TcUF0oqQiq3P3rwuWRZ","target":"5c351400","id":"bd27ab2c-8b59-4486-94f2-fd0b49d173a0"},"status":"OK"}}` + "\n"
	RESULT_JOB_STR_4 = `8dac154677f0b053b4fdf2a18fba93e45a0009805a37cd108f0fcee00b0d0600`
	TEST_JOB_STR_5   = `{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"73292077-4cb3-4d26-80ec-93d3805c448f","job":{"blob":"0707efeccdd405566a7baee167953d5d3754d82087011228d2bdb36ac0b2798ebb8a264ada03aa00000000041a4cfc113c9be9ce1d7214298837b250a5ea70396dd166681d6dbaa1eb9b2703","job_id":"FF0V3jespRJxKE+aNV/rAFXL6YXu","target":"9bc42000","id":"73292077-4cb3-4d26-80ec-93d3805c448f"},"status":"OK"}}` + "\n"
	RESULT_JOB_STR_5 = `46cfea7d7afe4739d517783e3b4b78a48fe173288cd2926a774cc71c94881b00`
	TEST_JOB_STR     = TEST_JOB_STR_5
	RESULT_JOB_STR   = RESULT_JOB_STR_5
)

func NewTestServer(port int) (*TestServer, error) {
	snl, err := stoppablenetlistener.New(port)
	if err != nil {
		return nil, err
	}
	log.Infof("Listening on %v", snl.Addr())
	ts := &TestServer{
		snl,
		make(chan *ClientRequest),
	}
	go func() {
		for {
			conn, err := snl.Accept()
			if err != nil {
				break
			}
			log.Infof("Received connection from '%v'", conn.RemoteAddr())
			go ts.handleConnection(conn)
		}
	}()
	return ts, nil
}

func (ts *TestServer) handleConnection(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		msg = strings.TrimSpace(msg)
		if err != nil {
			log.Infof("Breaking connection from IP '%v': %v", conn.RemoteAddr(), err)
			break
		}
		// log.Infof("Received message from IP: %v: %v", conn.RemoteAddr(), msg)
		var request Request
		if err := json.Unmarshal([]byte(msg), &request); err != nil {
			log.Errorf("Message not in JSON format: %v", err)
		} else {
			ts.RequestChan <- &ClientRequest{
				conn,
				&request,
			}
		}
	}
}
