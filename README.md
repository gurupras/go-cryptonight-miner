# go-cryptonight-miner

A simple Cryptonight miner written in Golang with CGO wrappers for the core cryptonight hashing functions. This miner is draws a lot of its inspiration (code) from the xmrig miner and thus exhibits similar performance.

The current implementation is more of a working prototype than a high-performant, full-fledged, cross-platform miner. Thus, it does not offer a lot of the flexibility like other projects.
I'm hoping that this can be addressed via PRs by developers far more knowledgeable than me.

## What does it do?
  - Working CPU miner
  - Working AMD GPU miner

## Mining Performance
CPU mining is slightly slower than xmrig while GPU mining varies quite a bit.

### CPU performance
**Intel i7 4790K @ 4.4GHz**  

  - xmrig:                 289H/s
  - go-cryptonight-miner:  279H/s

### GPU performance
**RX Vega64 @ 1458+1150MHz**  

The results reported below are averages over 15 minutes. The GPU was reset between measurements. These results are to be taken with a pinch of salt as the hashrate reporting logic may be incorrect.

**Configuration**: Threads: 2, Intensity: 1856/1600, Worksize: 8
  - xmrig-amd:            2022H/s
  - go-cryptonight-miner: 2024H/s
  
**Configuration**: Threads: 1, Intensity: 1600, Worksize: 8
  - xmrig-amd:            1453H/s
  - go-cryptonight-miner: 1451H/s

**Note**: Due to unknown reasons, in rare cases, the GPU hashrate has sometimes been observed to be ~150H/s lower than xmrig. Performing a GPU reset seems to fix this.

# Build Instructions

## Windows
The MSYS2 platform is required to build the miner from source.  

After installing the MSYS2 platform, use the following instructions for a 32/64-bit build.

### MSYS2 64-bit
Open the `mingw64.exe` shell and run:

    pacman -Sy
    pacman -S mingw-w64-x86_64-gcc

### MSYS2 32-bit
Open the `mingw32.exe` shell and run:

    pacman -Sy
    pacman -S mingw-w64-i686-gcc

### Building the CPU miner
Navigate to `cmd/cpuminer` and run `go build`

### Building the AMD GPU miner
The GPU miner requires the OpenCL libraries and headers to compile successfully.

For AMD GPUs, this requires installation of the AMD APP SDK. Once the APP SDK has been installed, the GPU miner can be built with the following commands:

    cd cmd/amd-miner
    go get
    export  CGO_CFLAGS=-I<path-to-AMD_APP_SDK>/include/
    export  CGO_LDFLAGS=-L<path-to-AMD_APP_SDK>/3.0/lib/x86_64/
    go install -tags="cl11" github.com/rainliu/gocl/cl    # This speeds up future builds
    go build -tags="cl11"
    
