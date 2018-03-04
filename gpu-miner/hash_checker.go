package gpuminer

import (
	"github.com/gurupras/go-cryptonight-miner/cpu-miner/xmrig_crypto"
	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
)

type HashResult struct {
	id uint32
	*stratum.StratumContext
	*xmrig_crypto.XMRigWork
}

var (
	HashCheckChan chan *HashResult = make(chan *HashResult, 256)
)

func RunHashChecker() {
	globalMem, err := xmrig_crypto.SetupHugePages(1)
	if err != nil {
		log.Fatalf("Failed to initialize hugepages: %v", err)
	}
	ctx, err := xmrig_crypto.SetupCryptonightContext(globalMem, 0)
	if err != nil {
		log.Fatalf("Failed to intialize context: %v", err)
	}

	for hr := range HashCheckChan {
		if hashBytes, foundHash := xmrig_crypto.CryptonightHash(hr.XMRigWork, ctx); foundHash {
			hashHex, err := stratum.BinToHex(hashBytes)
			if err != nil {
				log.Errorf("RunHashChecker: Failed to convert hash bytes to hex: %v", err)
				continue
			}
			log.Debugf("Submitting id=%d job=%v result=%v", hr.id, hr.XMRigWork.Work.JobID, hashHex)
			hr.SubmitWork(hr.XMRigWork.Work, hashHex)
		} else {
			log.Errorf("GPU #%d COMPUTE ERROR", hr.id)
		}
	}
}
