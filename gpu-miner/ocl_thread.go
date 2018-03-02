package gpuminer

import (
	"bytes"
	"encoding/binary"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	stratum "github.com/gurupras/go-stratum-client"
	cpuminer "github.com/gurupras/go-stratum-client/cpu-miner"
	"github.com/gurupras/go-stratum-client/cpu-miner/xmrig_crypto"
	amdgpu "github.com/gurupras/go-stratum-client/gpu-miner/amd"
	"github.com/gurupras/go-stratum-client/miner"
	"github.com/prometheus/common/log"
	"github.com/rainliu/gocl/cl"
)

var (
	TotalMiners uint32 = 0
	minerId     uint32 = 0
)

type GPUMiner struct {
	*stratum.StratumContext
	*miner.Miner
	Context   *amdgpu.GPUContext
	Index     int
	Intensity int
	WorkSize  int
}

func NewGPUMiner(sc *stratum.StratumContext, index, intensity, worksize int) *GPUMiner {
	miner := &GPUMiner{
		sc,
		miner.New(minerId),
		amdgpu.New(index, intensity, worksize),
		index,
		intensity,
		worksize,
	}
	atomic.AddUint32(&TotalMiners, 1)
	atomic.AddUint32(&minerId, 1)
	return miner
}

type CLResult []cl.CL_int

func (clr CLResult) Bytes() []byte {
	var dummy cl.CL_int
	ret := make([]byte, len(clr)*int(unsafe.Sizeof(dummy)))
	buf := bytes.NewBuffer(ret)

	b := make([]byte, 4)
	for _, v := range clr {
		binary.LittleEndian.PutUint32(b, uint32(v))
		buf.Write(b)
	}
	return ret
}

func (clr CLResult) Zero() {
	for i := 0; i < len(clr); i++ {
		clr[i] = 0x0
	}
}

func (m *GPUMiner) Run() error {
	results := make(CLResult, 0x100)

	defaultNonce := 0xffffffff / int(TotalMiners) * (int(m.Id()))
	workLock := sync.Mutex{}
	work := xmrig_crypto.NewXMRigWork()
	var newWork *stratum.Work

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

	noncePtr := work.NoncePtr

	consumeWork := func() {
		workLock.Lock()
		defer workLock.Unlock()
		if newWork == nil || newWork.JobID == work.JobID {
			return
		}
		//log.Debugf("Thread-%d: Got new work - %s", m.id, newWork.JobID)
		//log.Debugf("Thread-%d: blob: %v", stratum.BinToStr(newWork.Data))
		cpuminer.WorkCopy(work.Work, newWork)
		work.UpdateCData()
		*noncePtr = uint32(defaultNonce)
		amdgpu.XMRSetWork(m.Context, work.Data, work.Size, work.Target)
	}

	var (
		startTime time.Time
		endTime   time.Time
	)
	startTime = time.Now()

	initialWg.Wait()
	log.Debugf("Got first job")
	consumeWork()

	for {
		results.Zero()

		amdgpu.XMRRunWork(m.Context, results)

		for i := 0; i < int(results[0xFF]); i++ {
			*noncePtr = uint32(results[i])
			m.SubmitWork(work)
		}

		endTime = time.Now()
		diff := endTime.Sub(startTime)
		startTime = endTime
		m.InformHashrate(uint32(m.Context.RawIntensity), diff)

		consumeWork()
	}
}

// We need to check the hash. So just send the work down on HashCheckChan
func (m *GPUMiner) SubmitWork(work *xmrig_crypto.XMRigWork) error {
	hashResult := &HashResult{
		m.Id(),
		m.StratumContext,
		work,
	}
	HashCheckChan <- hashResult
	return nil
}
