package cpuminer

import (
	"strings"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/fatih/set"
	stratum "github.com/gurupras/go-stratum-client"
)

var (
	TotalMiners uint32 = 0
	minerId     uint32 = 0
)

type HashRate struct {
	Hashes    uint32
	TimeTaken time.Duration
}

type CPUMiner struct {
	*stratum.StratumContext
	CryptonightContext unsafe.Pointer
	id                 uint32
	hashrateListeners  set.Interface
}

type Interface interface {
	Id() uint32
	Run() error
	RegisterHashrateListener(chan *HashRate)
}

func New(sc *stratum.StratumContext) *CPUMiner {
	miner := &CPUMiner{
		sc,
		nil,
		minerId,
		set.New(),
	}
	atomic.AddUint32(&minerId, 1)
	atomic.AddUint32(&TotalMiners, 1)
	return miner
}

func (m *CPUMiner) Id() uint32 {
	return m.id
}

func (m *CPUMiner) RegisterHashrateListener(hrChan chan *HashRate) {
	m.hashrateListeners.Add(hrChan)
}

func (m *CPUMiner) InformHashrate(hashes uint32, timeTaken time.Duration) {
	data := &HashRate{
		hashes,
		timeTaken,
	}
	for _, obj := range m.hashrateListeners.List() {
		hrChan := obj.(chan *HashRate)
		hrChan <- data
	}
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
