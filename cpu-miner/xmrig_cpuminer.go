package cpuminer

/*
#cgo CFLAGS: -I. -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native
#include "crypto/helpers.h"
*/
import "C"
import (
	"sync"
	"time"

	stratum "github.com/gurupras/go-stratum-client"
	"github.com/gurupras/go-stratum-client/cpu-miner/xmrig_crypto"
	log "github.com/sirupsen/logrus"
)

type XMRigCPUMiner struct {
	*CPUMiner
}

func NewXMRigCPUMiner(sc *stratum.StratumContext) Interface {
	miner := New(sc)
	return &XMRigCPUMiner{
		miner,
	}
}

func (m *XMRigCPUMiner) Run() error {
	defaultNonce := 0xffffffff / int(TotalMiners) * (int(m.id))
	workLock := sync.Mutex{}
	work := stratum.NewWork()
	var newWork *stratum.Work
	var err error

	workChan := make(chan *stratum.Work, 0)

	m.StratumContext.RegisterWorkListener(workChan)
	go func() {
		for work := range workChan {
			workLock.Lock()
			newWork = work
			workLock.Unlock()
			log.Debugf("miner-new-work: Updated work - %s", newWork.JobID)
		}
	}()

	if m.CryptonightContext, err = xmrig_crypto.SetupCryptonightContext(); err != nil {
		return err
	}

	noncePtr := work.NoncePtr

	consumeWork := func() {
		if newWork == work {
			return
		}
		log.Debugf("Thread-%d: Got new work - %s", m.id, newWork.JobID)
		workLock.Lock()
		defer workLock.Unlock()
		workCopy(work, newWork)
		*noncePtr = uint32(defaultNonce)
	}

	for {
		var (
			hashesDone uint64 = 0
			startTime  time.Duration
			endTime    time.Duration
		)
		_ = hashesDone
		_ = startTime
		_ = endTime

		*noncePtr++

		if hashBytes, found := xmrig_crypto.CryptonightHash(work, m.CryptonightContext); found {
			m.SubmitWork(work, hashBytes)
		}
		consumeWork()
	}
	return nil
}

func (m *XMRigCPUMiner) SubmitWork(work *stratum.Work, hashBytes []byte) error {
	hashHex, err := stratum.BinToHex(hashBytes)
	if err != nil {
		return err
	}
	return m.StratumContext.SubmitWork(work, hashHex)
}
