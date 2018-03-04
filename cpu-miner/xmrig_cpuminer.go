package cpuminer

import (
	"strings"
	"sync"
	"unsafe"

	"github.com/gurupras/go-cryptonight-miner/cpu-miner/xmrig_crypto"
	"github.com/gurupras/go-cryptonight-miner/miner"
	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
)

var globalMemoryLock sync.Mutex
var globalMemory unsafe.Pointer

type XMRigCPUMiner struct {
	*CPUMiner
}

func NewXMRigCPUMiner(sc *stratum.StratumContext) miner.Interface {
	miner := New(sc)
	return &XMRigCPUMiner{
		miner,
	}
}

func (m *XMRigCPUMiner) Run() error {
	defaultNonce := 0xffffffff / int(TotalMiners) * (int(m.Id()))
	workLock := sync.Mutex{}
	work := xmrig_crypto.NewXMRigWork()
	var newWork *stratum.Work
	var err error

	globalMemoryLock.Lock()
	if globalMemory == nil {
		if globalMemory, err = xmrig_crypto.SetupHugePages(TotalMiners); err != nil {
			log.Fatalf("Failed to allocate hugepages: %v", err)
		}
	}
	globalMemoryLock.Unlock()

	workChan := make(chan *stratum.Work, 0)

	initialWg := sync.WaitGroup{}
	initialWg.Add(1)
	gotFirstJob := false

	m.StratumContext.RegisterWorkListener(workChan)
	go func() {
		for work := range workChan {
			workLock.Lock()
			newWork = work
			if !gotFirstJob {
				gotFirstJob = true
				initialWg.Done()
			}
			workLock.Unlock()
			log.Debugf("miner-new-work: Updated work - %s", newWork.JobID)
			log.Debugf("miner-new-work: target=%X", newWork.Target)
		}
	}()

	if m.CryptonightContext, err = xmrig_crypto.SetupCryptonightContext(globalMemory, m.Id()); err != nil {
		return err
	}

	noncePtr := work.NoncePtr

	consumeWork := func() {
		workLock.Lock()
		defer workLock.Unlock()
		if newWork == nil || strings.Compare(newWork.JobID, work.JobID) == 0 {
			return
		}
		//log.Debugf("Thread-%d: Got new work - %s", m.id, newWork.JobID)
		//log.Debugf("Thread-%d: blob: %v", stratum.BinToStr(newWork.Data))
		stratum.WorkCopy(work.Work, newWork)
		work.UpdateCData()
		*noncePtr = uint32(defaultNonce)
	}

	var (
		hashesDone uint32 = 0
	)

	initialWg.Wait()
	log.Debugf("Got first job")
	consumeWork()

	for {
		*noncePtr++
		hashesDone++

		if hashesDone&0xFF != 0 {
			m.InformHashrate(hashesDone)
			hashesDone = 0
		}

		if hashBytes, found := xmrig_crypto.CryptonightHash(work, m.CryptonightContext); found {
			m.SubmitWork(work, hashBytes)
		}
		consumeWork()
	}
	return nil
}

func (m *XMRigCPUMiner) SubmitWork(work *xmrig_crypto.XMRigWork, hashBytes []byte) error {
	hashHex, err := stratum.BinToHex(hashBytes)
	if err != nil {
		return err
	}
	return m.StratumContext.SubmitWork(work.Work, hashHex)
}
