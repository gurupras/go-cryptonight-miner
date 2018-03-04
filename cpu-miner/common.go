package cpuminer

import (
	"sync/atomic"
	"unsafe"

	"github.com/gurupras/go-cryptonite-miner/miner"
	stratum "github.com/gurupras/go-stratum-client"
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
