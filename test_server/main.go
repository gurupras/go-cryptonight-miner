package main

import (
	"strings"
	"sync"

	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.DebugLevel)
	server, err := stratum.NewTestServer(8888)
	if err != nil {
		log.Fatalf("%v")
	}
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for clientRequest := range server.RequestChan {
			if strings.Compare(clientRequest.Request.RemoteMethod, "login") == 0 || strings.Compare(clientRequest.Request.RemoteMethod, "submit") == 0 {
				if _, err := clientRequest.Conn.Write([]byte(stratum.TEST_JOB_STR_5)); err != nil {
					log.Errorf("Failed to send client test job: %v", err)
				}
			}
			log.Debugf("Received message: %v", clientRequest.Request)
		}
	}()
	wg.Wait()
}
