package stratum

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

type Work struct {
	Data       WorkData
	Target     []uint32 `json:"target"`
	JobID      string   `json:"job_id"`
	NoncePtr   *uint32
	Difficulty float64 `json:"difficulty"`
	XNonce2    string
}

func NewWork() *Work {
	work := &Work{}
	work.Data = make([]byte, 128)
	work.Target = make([]uint32, 8)
	work.NoncePtr = (*uint32)(unsafe.Pointer(&work.Data[39]))
	return work
}

type WorkData []byte

func (w WorkData) AsUint32Slice() []uint32 {
	ret := make([]uint32, 32)
	for i := 0; i < 32; i++ {
		val := binary.LittleEndian.Uint32(w[i*4 : i*4+4])
		// log.Debugf("bytes: %v  val[%d]=%d", data[i*4:i*4+4], i, val)
		ret[i] = val
	}
	return ret
}

func ParseWorkFromResponse(r *Response) (*Work, error) {
	result := r.Result
	if job, ok := result["job"]; !ok {
		return nil, fmt.Errorf("No job found")
	} else {
		return ParseWork(job.(map[string]interface{}))
	}
}

func ParseWork(args map[string]interface{}) (*Work, error) {
	jobId := args["job_id"].(string)
	hexBlob := args["blob"].(string)

	log.Debugf("job_id: %v", jobId)
	log.Debugf("hexblob: %v", hexBlob)
	blobLen := len(hexBlob)
	log.Debugf("blobLen: %v", blobLen)

	if blobLen%2 != 0 || ((blobLen/2) < 40 && blobLen != 0) || (blobLen/2) > 128 {
		return nil, fmt.Errorf("JSON invalid blob length")
	}

	if blobLen == 0 {
		return nil, fmt.Errorf("Blob length was 0?")
	}

	// TODO: Should there be a lock here?
	blob, err := HexToBin(hexBlob, blobLen/2)
	if err != nil {
		return nil, err
	}

	targetStr := args["target"].(string)
	log.Debugf("targetStr: %v", targetStr)
	b, err := HexToBin(targetStr, 4)
	target := binary.LittleEndian.Uint32(b)
	log.Debugf("target: %v", target)
	difficulty := float64(0xffffffff) / float64(target)
	log.Infof("Pool set difficulty: %.2f", difficulty)

	work := NewWork()

	copy(work.Data, blob)
	// XXX: Do we need to do this?
	for i := len(blob); i < len(work.Data); i++ {
		work.Data[i] = '\x00'
	}

	for i := 0; i < len(work.Target); i++ {
		work.Target[i] = 0xff
	}

	work.JobID = jobId
	work.Target[7] = target
	work.Difficulty = difficulty
	return work, nil

}
