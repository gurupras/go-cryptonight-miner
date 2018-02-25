package cpuminer

import (
	"fmt"
	"strings"
	"sync"
	"testing"

	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

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
	miner := NewCPUMinerMulti(st)
	go miner.Run()

	err = st.Connect(fmt.Sprintf("localhost:%d", port))
	require.Nil(err)
	err = st.Authorize("x", "y")
	require.Nil(err)

	wg.Wait()
}

func TestCPUMinerMulti(t *testing.T) {
	testCPUMiner(t, 4, NewCPUMinerMulti)
}
