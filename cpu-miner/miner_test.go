package cpuminer

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

var testConfig map[string]interface{}

func TestCryptonightSolver(t *testing.T) {
	require := require.New(t)
	log.SetLevel(log.DebugLevel)

	port := 8813
	server, err := stratum.NewTestServer(port)
	require.Nil(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for clientRequest := range server.RequestChan {
			if strings.Compare(clientRequest.Request.RemoteMethod, "login") == 0 {
				if _, err := clientRequest.Conn.Write([]byte(stratum.TEST_JOB_STR)); err != nil {
					log.Errorf("Failed to send client test job: %v", err)
				}
			} else {
				log.Infof("Received request: %v", clientRequest.Request)
			}
		}
	}()

	st := stratum.New()
	miner := New(st)
	go miner.Run()

	err = st.Connect(fmt.Sprintf("localhost:%d", port))
	require.Nil(err)
	err = st.Authorize("x", "y")
	require.Nil(err)

	wg.Wait()
}

func TestMiner(t *testing.T) {
	require := require.New(t)
	log.SetLevel(log.DebugLevel)

	sc := stratum.New()

	wg := sync.WaitGroup{}
	wg.Add(1)

	hashrateChan := make(chan *HashRate)
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

	numMiners := 4
	miners := make([]*CPUMiner, numMiners)
	for i := 0; i < numMiners; i++ {
		miner := New(sc)
		miner.RegisterHashrateListener(hashrateChan)
		miners[i] = miner
	}

	for i := 0; i < numMiners; i++ {
		go miners[i].Run()
	}

	responseChan := make(chan *stratum.Response)
	validShares := 3
	go func() {
		response := <-responseChan
		_ = response
		validShares--
		if validShares == 0 {
			wg.Done()
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
	log.SetLevel(log.WarnLevel)

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
