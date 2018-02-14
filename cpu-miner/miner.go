package cpuminer

import (
	"encoding/binary"
	"sync/atomic"
	"time"

	stratum "github.com/gurupras/go-stratum-client"
)

var (
	TotalMiners uint32 = 1
	minerId     uint32 = 0
)

type CPUMiner struct {
	*stratum.StratumContext
	id uint32
}

func New(sc *stratum.StratumContext) *CPUMiner {
	id := atomic.AddUint32(&minerId, 1)
	return &CPUMiner{
		sc,
		id,
	}
}

func (m *CPUMiner) Run() error {
	var (
		maxNonce   uint32
		endNonce   uint32 = (0xffffffff / (uint32(TotalMiners * uint32(minerId+1)))) - 0x20
		scratchBuf []byte
		s          [16]byte
		i          int
		work       stratum.Work
	)

	nonce := 0
	// TODO: Add hugepages logic here
	for newWork := range m.WorkChan {
		var (
			hashesDone               uint32
			startTime, endTime, diff time.Duration
			max64                    int64
			rc                       int
		)

		copy(work.DataBytes[:39], newWork.DataBytes[:39])
		copy(work.DataBytes[43:], newWork.DataBytes[43:])
		binary.LittleEndian.PutUint32(work.DataBytes[39:43], 0xffffffff/TotalMiners*m.id)

		max64 = LP_SCANTIME
		if nonce+max64 > endNonce {
			maxNonce = endNonce
		} else {
			maxNonce = nonce + max64
		}

		hashesDone = 0
		startTime := time.Now()
		rc = scanhash_cryptonight(m.id, work.Data, work.Target, maxNonce, &hashesDone)
	}
	return nil
}
