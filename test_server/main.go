package main

import (
	"bufio"
	"encoding/json"
	"net"
	"strings"

	log "github.com/sirupsen/logrus"
)

func handleConnection(conn net.Conn) {
	for {
		msg, err := bufio.NewReader(conn).ReadString('\n')
		msg = strings.TrimSpace(msg)
		if err != nil {
			log.Infof("Breaking connection from IP '%v': %v", conn.RemoteAddr(), err)
			break
		}
		log.Infof("Received message from IP: %v: %v", conn.RemoteAddr(), msg)
		var request map[string]interface{}
		if err := json.Unmarshal([]byte(msg), &request); err != nil {
			// log.Errorf("Message not in JSON format: %v", err)
		} else {
		}
		// conn.Write([]byte(`{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"d568505d-2fe4-44df-8322-cb91bf6546e3","job":{"blob":"01019ecf8dd405cab6225b29c0df65593cc7a05818b5a52b5e20ec7e6eba7898e01f8406116767000000003968df2b39eaac10f75f15b62ccc2950031803973547a08f101d3878d993139e28","job_id":"SDZflakRiUH2mpZOZKWPPqXKVu/1","target":"8b4f0100","id":"d568505d-2fe4-44df-8322-cb91bf6546e3"},"status":"OK"}}` + "\n"))
		conn.Write([]byte(`{"id":1,"jsonrpc":"2.0","error":null,"result":{"id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335","job":{"blob":"0101c1ef8dd405b9a2d3278fbc35ba876422c82a94dd7befec695f42848d55cab06a96e31a34b300000000b1940501101bd2f22938ce55f25ddf3c584ab6915453073f3c09a5967df6966204","job_id":"Js5ps3OcKxJUCiEjtIz54ImuNMmA","target":"b88d0600","id":"8bc409ea-7ca7-4073-a596-31b1c8fb9335"},"status":"OK"}}
      `))
	}
}

func main() {
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		log.Fatalf("Failed to listen for connections: %v", err)
	}
	log.Infof("Listening on %v", l.Addr())
	for {
		conn, _ := l.Accept()
		log.Infof("Received connection from '%v'", conn.RemoteAddr())
		go handleConnection(conn)
	}
}
