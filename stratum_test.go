package stratum

import (
	"io/ioutil"
	"os"
	"sync"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

var testConfig map[string]interface{}

func connect(sc *StratumContext) error {
	err := sc.Connect(testConfig["pool"].(string))
	if err != nil {
		log.Debugf("Connected to pool..")
	}
	return err
}

func TestConnect(t *testing.T) {
	require := require.New(t)

	sc := New()
	err := connect(sc)
	require.Nil(err)
}

func TestBadAuthorize(t *testing.T) {
	require := require.New(t)

	sc := New()
	err := connect(sc)
	require.Nil(err)

	err = sc.Authorize("", testConfig["pass"].(string))
	require.NotNil(err)
}

func TestAuthorize(t *testing.T) {
	require := require.New(t)

	sc := New()
	err := connect(sc)
	require.Nil(err)

	err = sc.Authorize(testConfig["username"].(string), testConfig["pass"].(string))
	require.Nil(err)
}

func TestParseWork(t *testing.T) {
	// t.Skip("Usually skip since this is an infinite loop test")
	require := require.New(t)

	sc := New()
	err := connect(sc)
	require.Nil(err)

	err = sc.Authorize(testConfig["username"].(string), testConfig["pass"].(string))
	require.Nil(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for work := range sc.WorkChan {
			log.Infof("Got new work - difficulty: %d", work.Difficulty)
		}
	}()
	wg.Wait()
}

func TestMain(m *testing.M) {
	log.SetLevel(log.DebugLevel)

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
