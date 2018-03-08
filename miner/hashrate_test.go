package miner

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	colorable "github.com/mattn/go-colorable"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func generateHashRate(targetHashesPerSec int, delay time.Duration, signal <-chan struct{}, outChan chan<- *HashRate) {
	// First hashrate is 0
	// outChan <- &HashRate{
	// 	0,
	// 	time.Now(),
	// }
	// We will try and send 10 HashRates every delay
	hashesPerEntry := uint32(targetHashesPerSec) / 10
	for {
		select {
		case <-signal:
			return
		case <-time.After(delay / 10):
			outChan <- &HashRate{
				hashesPerEntry,
				time.Now(),
			}
		}
	}
}

func generateFixedHashRate(targetHashes int, delay time.Duration, signal <-chan struct{}, outChan chan<- *HashRate) {
	// First hashrate is 0
	// outChan <- &HashRate{
	// 	0,
	// 	time.Now(),
	// }
	for {
		select {
		case <-signal:
			return
		case <-time.After(delay):
			outChan <- &HashRate{
				uint32(targetHashes),
				time.Now(),
			}
		}
	}
}

func TestFixedHashRate(t *testing.T) {
	require := require.New(t)

	fixedHashRate := 2000

	signal := make(chan struct{})

	hrChan := make(chan *HashRate)
	outChan := make(chan HashRateTrackerArray)
	trackerDurations := []time.Duration{10 * time.Second, 15 * time.Second, 30 * time.Second}
	expectedResult := 0x7 // XXX: Change this if you change trackerDurations

	wg := sync.WaitGroup{}
	wg.Add(1)
	count := 10

	go SetupHashRateTrackers(5*time.Second, trackerDurations, hrChan, outChan)

	go func() {
		defer wg.Done()
		// All trackers have to report successful hashrate
		result := 0
		for array := range outChan {
			log.Infof(array.String())
			for idx, hrt := range array {
				avg := int(hrt.Average())
				if avg == 0 {
					require.True(hrt.durationDiff() < hrt.duration)
					continue
				}

				diff := uint32(math.Abs(float64(avg - fixedHashRate)))
				ratio := float64(diff) / float64(fixedHashRate)
				if ratio > 0.05 {
					msg := fmt.Sprintf("hrt(%s) average (%v) ~= expected (%v) diff=%d ratio=%.2f", hrt.DurationString(), avg, fixedHashRate, diff, ratio)
					log.Errorf(msg)
					log.Errorf("hashes: %v", hrt.Hashes())
					log.Errorf("times:  %v", hrt.Times())
					require.Fail(msg)
				}
				result |= (1 << uint(idx))
			}
			count--
			if count == 0 {
				break
			}
		}
		// We have 3 trackers and so lowest 3 bits need to be flipped to 1 ==> 0x7
		require.Equal(expectedResult, result)
	}()

	// go generateHashRate(fixedHashRate, 1*time.Second, signal, hrChan)
	go generateFixedHashRate(fixedHashRate, 1*time.Second, signal, hrChan)
	wg.Wait()
	signal <- struct{}{}
}

func TestMain(m *testing.M) {
	log.SetLevel(log.InfoLevel)

	if runtime.GOOS == "windows" {
		log.SetFormatter(&log.TextFormatter{ForceColors: true})
		log.SetOutput(colorable.NewColorableStdout())
	}

	os.Exit(m.Run())
}
