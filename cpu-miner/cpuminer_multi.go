package cpuminer

/*
#cgo CFLAGS: -I. -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native
#include "crypto/helpers.h"
*/
import "C"
import (
	"bytes"
	"sync"
	"time"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
	"github.com/gurupras/go-stratum-client/cpu-miner/crypto"
	log "github.com/sirupsen/logrus"
)

const (
	LP_SCANTIME = 30
)

type CPUMinerMulti struct {
	*CPUMiner
}

func NewCPUMinerMulti(sc *stratum.StratumContext) Interface {
	miner := New(sc)
	return &CPUMinerMulti{
		miner,
	}
}

func (m *CPUMinerMulti) Run() error {
	defaultNonce := 0xffffffff/int(TotalMiners)*(int(m.id)+1) - 0x20
	var (
		maxNonce uint32
		endNonce uint32 = uint32(defaultNonce)
		err      error
		restart  int = 0
	)

	workLock := sync.Mutex{}
	work := stratum.NewWork()
	var newWork *stratum.Work

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

	if m.CryptonightContext, err = crypto.SetupCryptonightContext(); err != nil {
		return err
	}

	noncePtr := work.NoncePtr
	restartPtr := unsafe.Pointer(&restart)

	for {
		var (
			hashesDone         uint32
			startTime, endTime time.Time
			diff               time.Duration
			max64              int64
		)
		hashesDonePtr := unsafe.Pointer(&hashesDone)

		// log.Debugf("nonce=%d endNonce=%d", *noncePtr, endNonce)
		// log.Debugf("work=%p newWork=%p", work, newWork)

		workLock.Lock()
		if *noncePtr >= endNonce && (bytes.Compare(work.Data[:39], newWork.Data[:39]) == 0 && bytes.Compare(work.Data[43:76], newWork.Data[43:76]) == 0) {
			log.Debugf("Thread-%d: Reset work", m.id)
			workCopy(work, newWork)
		}
		if newWork != nil && (bytes.Compare(work.Data[:39], newWork.Data[:39]) != 0 || bytes.Compare(work.Data[43:76], newWork.Data[43:76]) != 0) {
			// TODO: stratum_gen_work()?
			// Think this means we just reset the work
			workCopy(work, newWork)
			endNonce = uint32(defaultNonce)
			*noncePtr = uint32(0xffffffff) / TotalMiners * m.id
			log.Debugf("Thread-%d: Got new work - %s", m.id, work.JobID)
		} else {
			*noncePtr += 1
		}
		workLock.Unlock()

		max64 = LP_SCANTIME
		if int64(*noncePtr)+max64 > int64(endNonce) {
			maxNonce = endNonce
		} else {
			maxNonce = uint32(int64(*noncePtr) + max64)
		}
		// log.Debugf("max64=%d", max64)

		hashesDone = 0
		startTime = time.Now()
		// log.Debugf("Starting ScanHashCryptonight")
		foundNonce := crypto.ScanHashCryptonight(m.id, work, maxNonce, hashesDonePtr, m.CryptonightContext, restartPtr)
		endTime = time.Now()
		diff = endTime.Sub(startTime)
		m.InformHashrate(hashesDone, diff)
		// log.Debugf("Finished ScanHashCryptonight: %dms", diff.Nanoseconds()/1e6)
		if foundNonce {
			log.Debugf("Found nonce! submitting work...")
			if err = m.SubmitWork(work); err != nil {
				log.Errorf("Failed to submit work: %v", err)
			}
		}
		// log.Debugf("hashes done=%d", hashesDone)
	}
	return nil
}

func (m *CPUMinerMulti) SubmitWork(work *stratum.Work) error {
	hash := crypto.CryptonightHash(work.Data, len(work.Data))
	hashHex, err := stratum.BinToHex(hash)
	if err != nil {
		return err
	}
	return m.StratumContext.SubmitWork(work, hashHex)
}
