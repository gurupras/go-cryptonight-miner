package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/alecthomas/kingpin"
	cpuminer "github.com/gurupras/go-cryptonight-miner/cpu-miner"
	"github.com/gurupras/go-cryptonight-miner/miner"
	stratum "github.com/gurupras/go-stratum-client"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var (
	app        = kingpin.New("cpuminer", "CPU Cryptonight miner")
	config     = app.Flag("config-file", "YAML config file").Short('c').String()
	url        = app.Flag("url", "URL of the pool").Short('o').String()
	username   = app.Flag("username", "Username (usually the wallet address)").Short('u').String()
	password   = app.Flag("password", "Password").Short('p').Default("go-cryptonight-miner").String()
	threads    = app.Flag("threads", "Number of threads to run").Short('t').Default(fmt.Sprintf("%d", runtime.NumCPU())).Int()
	cpuprofile = app.Flag("cpuprofile", "Run CPU profiler").String()
	verbose    = app.Flag("verbose", "Enable verbose log messages").Short('v').Bool()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if *verbose {
		log.SetLevel(log.DebugLevel)
	}

	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatalf("Failed to create cpuprofile file: %v", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatalf("Failed to start CPU profile: %v", err)
		}
		log.Infof("Starting CPU profiling")
		defer pprof.StopCPUProfile()
	}

	// Start all logic here

	if len(*config) == 0 {
		if len(*url) == 0 || len(*username) == 0 {
			log.Fatalf("Must specify config or url, username and password")
		}
	} else {
		if len(*url) != 0 || len(*username) != 0 {
			log.Warningf("Using config over commandline arguments..url=%s username=%s, pass=%s", *url, *username, *password)
		}
	}
	var configData []byte
	if len(*config) != 0 {
		// Parse config file and extract necessary fields
		data, err := ioutil.ReadFile(*config)
		if err != nil {
			log.Fatalf("Failed to read config file: %v", err)
		}
		configData = data
	} else {
		minConfig := fmt.Sprintf(`
cpu_threads: %d
pools:
  - url: %v
    user: %v
    pass: %v
`, *threads, *url, *username, *password)
		configData = []byte(minConfig)
		log.Debugf("minConfig: %v", minConfig)
	}
	var config miner.Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse yaml into valid config: %v", err)
	}

	sc := stratum.New()

	hashrateChan := make(chan *miner.HashRate, 10)
	go miner.RunDefaultHashRateTrackers(hashrateChan)

	if config.CPUThreads == 0 {
		if *threads != 0 {
			config.CPUThreads = *threads
		} else {
			config.CPUThreads = runtime.NumCPU()
		}
	}

	numMiners := config.CPUThreads
	miners := make([]miner.Interface, numMiners)
	for i := 0; i < numMiners; i++ {
		miner := cpuminer.NewXMRigCPUMiner(sc)
		miner.RegisterHashrateListener(hashrateChan)
		miners[i] = miner
	}
	log.Infof("# Threads: %v", numMiners)

	wg := sync.WaitGroup{}
	wg.Add(1)
	for i := 0; i < numMiners; i++ {
		go miners[i].Run()
	}

	// responseChan := make(chan *stratum.Response)
	//
	// sc.RegisterResponseListener(responseChan)

	pool := config.Pools[0]
	if err := sc.Connect(pool.Url); err != nil {
		log.Fatalf("Failed to connect to url :%v  - %v", *url, err)
	}

	if err := sc.Authorize(pool.User, pool.Pass); err != nil {
		log.Fatalf("Failed to authorize with server: %v", err)
	}

	if *cpuprofile != "" {
		time.Sleep(300 * time.Second)
	} else {
		wg.Wait() // blocks forever
	}
}
