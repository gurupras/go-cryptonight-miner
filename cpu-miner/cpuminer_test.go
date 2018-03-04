package cpuminer

import (
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	stratum "github.com/gurupras/go-stratum-client"
	"github.com/gurupras/go-cryptonite-miner/miner"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

var testConfig map[string]interface{}

type constructor func(sc *stratum.StratumContext) miner.Interface

func testCPUMiner(t *testing.T, numMiners int, constructor constructor) {
	require := require.New(t)

	sc := stratum.New()

	wg := sync.WaitGroup{}
	wg.Add(1)

	hashrateChan := make(chan *miner.HashRate)
	go func() {
		duration := 10 * time.Second
		totalHashes := uint32(0)

		startTime := time.Now()
		for hr := range hashrateChan {
			now := time.Now()
			if now.Sub(startTime) < duration {
				totalHashes += hr.Hashes
			} else {
				log.Infof("Speed: %dH/s", uint32(float64(totalHashes)/(now.Sub(startTime).Seconds())))
				totalHashes = 0
				startTime = time.Now()
			}
		}
	}()

	miners := make([]miner.Interface, numMiners)
	for i := 0; i < numMiners; i++ {
		miner := constructor(sc)
		miner.RegisterHashrateListener(hashrateChan)
		miners[i] = miner
	}

	for i := 0; i < numMiners; i++ {
		go miners[i].Run()
	}

	responseChan := make(chan *stratum.Response)
	validShares := 3
	go func() {
		for response := range responseChan {
			if strings.Compare(response.Result["status"].(string), "OK") == 0 {
				validShares--
				if validShares == 0 {
					log.Debugf("Valid shares requirement met. Terminating test")
					wg.Done()
				}
			}
		}
	}()

	sc.RegisterResponseListener(responseChan)

	err := sc.Connect(testConfig["pool"].(string))
	require.Nil(err)

	err = sc.Authorize(testConfig["username"].(string), testConfig["pass"].(string))
	require.Nil(err)

	wg.Wait()
}

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)

	b, err := ioutil.ReadFile("test-config.yaml")
	if err != nil {
		log.Errorf("No test-config.yaml")
		str := `pool:
username:
pass:
`
		if err := ioutil.WriteFile("test-config.yaml", []byte(str), 0666); err != nil {
			log.Errorf("Failed to create test-config.yaml: %v", err)
		} else {
			log.Infof("Created test-config.yaml..run tests after filling it out")
			os.Exit(-1)
		}
	} else {
		if err := yaml.Unmarshal(b, &testConfig); err != nil {
			log.Fatalf("Failed to unmarshal test-config.yaml: %v", err)
		}
	}
	os.Exit(m.Run())
}
