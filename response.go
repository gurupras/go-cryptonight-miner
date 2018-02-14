package stratum

import (
	"encoding/json"
)

type Response struct {
	MessageID interface{}            `json:"id"`
	Result    map[string]interface{} `json:"result"`
	Error     *StratumError          `json:"error"`
}

func (r *Response) String() string {
	b, _ := json.Marshal(r)
	return string(b)
}
