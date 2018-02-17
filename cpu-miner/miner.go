package cpuminer

/*
#cgo CFLAGS: -I. -Wno-pointer-sign -Wno-pointer-to-int-cast -maes -march=native
#include "crypto/helpers.h"
*/
import "C"
import (
	"bytes"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/fatih/set"
	stratum "github.com/gurupras/go-stratum-client"
	"github.com/gurupras/go-stratum-client/cpu-miner/crypto"
	log "github.com/sirupsen/logrus"
)

const (
	LP_SCANTIME = 60
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

func (m *CPUMiner) Run() error {
	defaultNonce := 0xffffffff/int(TotalMiners)*(int(m.id)+1) - 0x20
	var (
		maxNonce uint32
		endNonce uint32 = uint32(defaultNonce)
		err      error
		restart  int = 0
	)

	workLock := sync.Mutex{}
	work := stratum.NewWork()
	newWork := m.StratumContext.Work

	workChan := make(chan *stratum.Work, 0)

	m.StratumContext.RegisterWorkListener(workChan)
	go func() {
		for work := range workChan {
			workLock.Lock()
			newWork = work
			workLock.Unlock()
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
		if m.StratumContext.LastSubmittedWork == work {
			if err := m.StratumContext.GetJob(); err != nil {
				log.Errorf("Failed to get job")
			}
		}
		if newWork != nil && (bytes.Compare(work.Data[:39], newWork.Data[:39]) != 0 || bytes.Compare(work.Data[43:76], newWork.Data[43:76]) != 0) {
			// TODO: stratum_gen_work()?
			// Think this means we just reset the work
			workCopy(work, newWork)
			endNonce = uint32(defaultNonce)
			*noncePtr = uint32(0xffffffff) / TotalMiners * m.id
			log.Debugf("Thread-%d: Got new work..starting nonce=%d", m.id, *noncePtr)
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

func (m *CPUMiner) SubmitWork(work *stratum.Work) error {
	hash := crypto.CryptonightHash(work.Data, len(work.Data))
	hashHex, err := stratum.BinToHex(hash)
	if err != nil {
		return err
	}
	return m.StratumContext.SubmitWork(work, hashHex)
}

func workCopy(dest *stratum.Work, src *stratum.Work) {
	copy(dest.Data, src.Data)
	copy(dest.Target, src.Target)
	if strings.Compare(src.JobID, "") != 0 {
		dest.JobID = src.JobID
	}
	if strings.Compare(src.XNonce2, "") != 0 {
		dest.XNonce2 = src.XNonce2
	}
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
