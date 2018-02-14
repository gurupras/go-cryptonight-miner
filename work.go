package stratum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"unsafe"

	log "github.com/sirupsen/logrus"
)

type Work struct {
	Data       [32]uint32 `json:"blob"`
	DataBytes  []byte
	Target     [8]uint32 `json:"target"`
	JobID      string    `json:"job_id"`
	Difficulty float64   `json:"difficulty"`
	XNonce2    string
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

	if blobLen != 0 {
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
		_ = jobId

		work := &Work{}

		blobReader := bytes.NewReader(blob)
		for i := 0; i < len(blob)/4; i++ {
			var val uint32
			if err := binary.Read(blobReader, binary.LittleEndian, val); err != nil {
				return nil, err
			}
			work.Data[i] = val
		}

		for i := 0; i < len(work.Target); i++ {
			work.Target[i] = 0xff
		}
		work.Target[7] = target
		work.Difficulty = difficulty
		work.DataBytes = *(*[]byte)(unsafe.Pointer(&work.Data))
		return work, nil
	}
	return nil, fmt.Errorf("Blob length was 0?")
}
