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

func TestWithExternalPool(t *testing.T) {
	t.Skip("Skipping external pool test")
	testCPUMiner(t, 4, NewXMRigCPUMiner)
}

func TestXMRigSolver(t *testing.T) {
	log.Warnf("This test may take a while depending on hashing rate")
	require := require.New(t)

	port := 44144
	server, err := stratum.NewTestServer(port)
	require.Nil(err)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		for clientRequest := range server.RequestChan {
			log.Debugf("server: Received message: %v", clientRequest.Request)
			if strings.Compare(clientRequest.Request.RemoteMethod, "login") == 0 {
				if _, err := clientRequest.Conn.Write([]byte(stratum.AUTH_RESPONSE_STR_4)); err != nil {
					log.Errorf("Failed to send client test job: %v", err)
				}
			} else if strings.Compare(clientRequest.Request.RemoteMethod, "submit") == 0 {
				params := clientRequest.Request.Parameters.(map[string]interface{})
				result := params["result"].(string)
				require.Equal(stratum.AUTH_RESPONSE_RESULT_4, result)
				defer server.StoppableNetListener.Stop()
				wg.Done()
			}
		}
	}()

	sc := stratum.New()

	miner := NewXMRigCPUMiner(sc)
	go miner.Run()

	sc.Connect(fmt.Sprintf("localhost:%d", port))
	sc.Authorize("test", "x")

	wg.Wait()
}
