package stratum

import "encoding/json"

type Request struct {
	MessageID    interface{} `json:"id"`
	RemoteMethod string      `json:"method"`
	Parameters   interface{} `json:"params"`
}

func NewRequest(id int, method string, args interface{}) *Request {
	return &Request{
		id,
		method,
		args,
	}
}

func (br *Request) JsonRPCString() (string, error) {
	payload := make(map[string]interface{})
	payload["jsonrpc"] = "2.0"
	payload["method"] = br.RemoteMethod
	payload["id"] = br.MessageID
	payload["params"] = br.Parameters

	b, err := json.Marshal(br)
	if err != nil {
		return "", err
	}
	return string(b) + "\n", nil

}
