package stratum

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

type Work struct {
	Data       WorkData
	Target     uint64 `json:"target"`
	JobID      string `json:"job_id"`
	NoncePtr   *uint32
	Difficulty float64 `json:"difficulty"`
	XNonce2    string
	Size       int
}

func NewWork() *Work {
	work := &Work{}
	work.Data = make([]byte, 84)
	work.Target = 0
	work.NoncePtr = (*uint32)(unsafe.Pointer(&work.Data[39]))
	return work
}

type WorkData []byte

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
	blob, err := HexToBin(hexBlob, blobLen)
	if err != nil {
		return nil, err
	}

	log.Debugf("blob bytes=%v", BinToStr(blob))

	targetStr := args["target"].(string)
	log.Debugf("targetStr: %v", targetStr)
	b, err := HexToBin(targetStr, 8)
	target := uint64(binary.LittleEndian.Uint32(b))
	target64 := math.MaxUint64 / (uint64(0xFFFFFFFF) / target)
	target = target64
	log.Debugf("target: %X", target)
	difficulty := float64(0xFFFFFFFFFFFFFFFF) / float64(target64)
	log.Infof("Pool set difficulty: %.2f", difficulty)

	work := NewWork()

	copy(work.Data, blob)
	// XXX: Do we need to do this?
	for i := len(blob); i < len(work.Data); i++ {
		work.Data[i] = '\x00'
	}

	work.Size = blobLen / 2
	work.JobID = jobId
	work.Target = target
	work.Difficulty = difficulty
	return work, nil
}
