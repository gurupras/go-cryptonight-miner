package cpuminer

import (
	"strings"
	"sync/atomic"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
	"github.com/gurupras/go-stratum-client/miner"
)

var (
	TotalMiners uint32 = 0
	minerId     uint32 = 0
)

type CPUMiner struct {
	*stratum.StratumContext
	*miner.Miner
	CryptonightContext unsafe.Pointer
}

func New(sc *stratum.StratumContext) *CPUMiner {
	miner := &CPUMiner{
		sc,
		miner.New(minerId),
		nil,
	}
	atomic.AddUint32(&minerId, 1)
	atomic.AddUint32(&TotalMiners, 1)
	return miner
}

func workCopy(dest *stratum.Work, src *stratum.Work) {
	copy(dest.Data, src.Data)
	dest.Size = src.Size
	dest.Difficulty = src.Difficulty
	dest.Target = src.Target
	if strings.Compare(src.JobID, "") != 0 {
		dest.JobID = src.JobID
	}
	if strings.Compare(src.XNonce2, "") != 0 {
		dest.XNonce2 = src.XNonce2
	}
}
