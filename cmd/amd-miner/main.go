package main

import (
	"io/ioutil"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/alecthomas/kingpin"
	gpuminer "github.com/gurupras/go-cryptonight-miner/gpu-miner"
	amdgpu "github.com/gurupras/go-cryptonight-miner/gpu-miner/amd"
	"github.com/gurupras/go-cryptonight-miner/gpu-miner/gpucontext"
	"github.com/gurupras/go-cryptonight-miner/miner"
	stratum "github.com/gurupras/go-stratum-client"
	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

var (
	app        = kingpin.New("cpuminer", "CPU Cryptonight miner")
	config     = app.Flag("config-file", "YAML config file").Short('c').Required().String()
	verbose    = app.Flag("verbose", "Enable verbose log messages").Short('v').Bool()
	debug      = app.Flag("debug", "Enable miner debugging log messages").Short('d').Default("false").Bool()
	useC       = app.Flag("use C", "Use C functions to intialize OpenCL  rather than Golang").Short('C').Default("false").Bool()
	cpuprofile = app.Flag("cpuprofile", "Run CPU profiler").String()
)

func main() {
	kingpin.MustParse(app.Parse(os.Args[1:]))

	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}

	amdgpu.UseC = *useC

	if *verbose {
		log.SetLevel(log.DebugLevel)
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

	// Parse config file and extract necessary fields
	configData, err := ioutil.ReadFile(*config)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	var config miner.Config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse yaml into valid config: %v", err)
	}

	sc := stratum.New()

	hashrateChan := make(chan *miner.HashRate, 10)
	go miner.SetupHashRateLogger(hashrateChan)

	numMiners := len(config.Threads)
	miners := make([]miner.Interface, numMiners)
	gpuContexts := make([]*gpucontext.GPUContext, numMiners)

	for i := 0; i < numMiners; i++ {
		threadInfo := config.Threads[i]
		miner := gpuminer.NewGPUMiner(sc, threadInfo.Index, threadInfo.Intensity, threadInfo.WorkSize)
		miner.RegisterHashrateListener(hashrateChan)
		gpuContexts[i] = miner.Context
		miners[i] = miner
		miner.SetDebug(*debug)
	}

	if err := amdgpu.InitOpenCL(gpuContexts, numMiners, config.OpenCLPlatform); err != nil {
		log.Fatalf("Failed to initialize OpenCL: %v", err)
	}

	go gpuminer.RunHashChecker()

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
		log.Fatalf("Failed to connect to url :%v  - %v", pool.Url, err)
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
