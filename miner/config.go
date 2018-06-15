package miner

// Config structure representing config JSON file
// Add any relevant fields here
// Config structure representing config JSON file
// Add any relevant fields here
type Config struct {
	Algorithm         string      `json:"algo" yaml:"algo"`
	Background        bool        `json:"background" yaml:"background"`
	Colors            bool        `json:"colors" yaml:"colors"`
	DonateLevel       float64     `json:"donate-level" yaml:"donate-level"`
	LogFile           *string     `json:"log-file" yaml:"log-file"`
	PrintTime         int         `json:"print-time" yaml:"print-time"`
	Retries           int         `json:"retries" yaml:"retries"`
	RetryPause        int         `json:"retry-pause" yaml:"retry-pause"`
	Syslog            bool        `json:"syslog" yaml:"syslog"`
	OpenCLPlatform    int         `json:"opencl-platform" yaml:"opencl-platform"`
	CPUThreads        int         `json:"cpu_threads" yaml:"cpu_threads"`
	DeviceInstanceIDs []string    `json:"device_instance_ids" yaml:"device_instance_ids"`
	Threads           []GPUThread `json:"threads" yaml:"threads"`
	Pools             []Pool      `json:"pools" yaml:"pools"`
	// Arguments to support miners like cpuminer-multi
	Url   string `json:"url" yaml:"url"`
	User  string `json:"user" yaml:"user"`
	Pass  string `json:"pass" yaml:"pass"`
	Proxy string `json:"proxy" yaml:"proxy"`
}

// GPUThread structure representing a GPU thread
type GPUThread struct {
	Index       int  `json:"index" yaml:"index"`
	DeviceIndex *int `json:"device_index" yaml:"device_index"`
	Intensity   int  `json:"intensity" yaml:"intensity"`
	WorkSize    int  `json:"worksize" yaml:"worksize"`
	AffineToCPU bool `json:"affine_to_cpu" yaml:"affine_to_cpu"`
}

// Pool structure representing a pool
type Pool struct {
	Url        string  `json:"url" yaml:"url"`
	User       string  `json:"user" yaml:"user"`
	Pass       string  `json:"pass" yaml:"pass"`
	Keepalive  bool    `json:"keepalive" yaml:"keepalive"`
	Nicehash   bool    `json:"nicehash" yaml:"nicehash"`
	Coin       *string `json:"coin" yaml:"coin"`
	PoolName   *string `json:"pool_name" yaml:"pool_name"`
	WalletName *string `json:"wallet_name" yaml:"wallet_name"`
	Label      *string `json:"label" yaml:"label"`
}
