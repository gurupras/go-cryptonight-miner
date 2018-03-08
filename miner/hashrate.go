package miner

import (
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	DefaultTrackerDurations = []time.Duration{15 * time.Second, 60 * time.Second, 15 * time.Minute}
)

type HashRate struct {
	Hashes uint32
	Time   time.Time
}

type HashRateTracker struct {
	hashRates []*HashRate
	duration  time.Duration
	hashes    uint32
	max       uint32
}

func NewHashRateTracker(duration time.Duration) *HashRateTracker {
	hrt := &HashRateTracker{}
	hrt.hashRates = make([]*HashRate, 0)
	hrt.duration = duration
	hrt.hashes = 0
	return hrt
}

func (hrt *HashRateTracker) Add(hr *HashRate) {
	hrt.hashRates = append(hrt.hashRates, hr)
	duration := hr.Time.Sub(hrt.hashRates[0].Time)
	if duration > hrt.duration*2 {
		hrt.hashes -= hrt.hashRates[0].Hashes
		hrt.hashRates = hrt.hashRates[1:]
	}
	hrt.hashes += hr.Hashes
}

func (hrt *HashRateTracker) durationDiff() time.Duration {
	size := len(hrt.hashRates)
	return hrt.hashRates[size-1].Time.Sub(hrt.hashRates[0].Time)
}

func (hrt *HashRateTracker) Average() uint32 {
	duration := hrt.durationDiff()
	if duration < hrt.duration {
		return 0
	}
	size := len(hrt.hashRates)

	totalHashes := hrt.hashRates[size-1].Hashes
	for i := size - 2; i > 0; i-- {
		totalHashes += hrt.hashRates[i].Hashes
		duration = hrt.hashRates[size-1].Time.Sub(hrt.hashRates[i-1].Time)
		if duration >= hrt.duration {
			break
		}
	}
	avg := uint32(float64(totalHashes) / duration.Seconds())
	if avg > hrt.max {
		hrt.max = avg
	}
	return avg
}

func (hrt *HashRateTracker) AverageAsString() string {
	avg := hrt.Average()
	if avg == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d", avg)
}

// Taken from https://stackoverflow.com/a/41336257/1761555
func shortDur(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = s[:len(s)-2]
	}
	if strings.HasSuffix(s, "h0m") {
		s = s[:len(s)-2]
	}
	return s
}

func (hrt *HashRateTracker) DurationString() string {
	return shortDur(hrt.duration)
}

func (hrt *HashRateTracker) Hashes() []uint32 {
	hashes := make([]uint32, len(hrt.hashRates))
	for idx, hr := range hrt.hashRates {
		hashes[idx] = hr.Hashes
	}
	return hashes
}

func (hrt *HashRateTracker) Times() []float64 {
	times := make([]float64, len(hrt.hashRates))
	zeroOffset := hrt.hashRates[0].Time
	for idx, hr := range hrt.hashRates {
		times[idx] = hr.Time.Sub(zeroOffset).Seconds()
	}
	return times
}

// HashRateTrackerArray is a wrapper for an array of HashRateTrackers.
// It implements the Stringer interface which provides an easy way
// to print out the hashrates of all trackers
type HashRateTrackerArray []*HashRateTracker

func (o HashRateTrackerArray) String() string {
	buf := make([]string, 0)
	buf = append(buf, fmt.Sprintf("\x1B[01;37mspeed"))

	durationStrings := make([]string, len(o))
	hashRateStrings := make([]string, len(o))
	maxHashRate := uint32(0)

	for idx, hrt := range o {
		durationStrings[idx] = hrt.DurationString()
		hashRateStrings[idx] = fmt.Sprintf("\x1B[01;36m%4s", hrt.AverageAsString())
		if hrt.max > maxHashRate {
			maxHashRate = hrt.max
		}
	}
	buf = append(buf, fmt.Sprintf("\x1B[0m %v", strings.Join(durationStrings, "/")))
	buf = append(buf, strings.Join(hashRateStrings, "\x1B[0m "))
	buf = append(buf, "\x1B[01;36mH/s")

	// max
	var maxHashRateStr string
	if maxHashRate == 0 {
		maxHashRateStr = "n/a"
	} else {
		maxHashRateStr = fmt.Sprintf("%v", maxHashRate)
	}
	buf = append(buf, fmt.Sprintf("\x1B[0m max: \x1B[01;36m%s H/s", maxHashRateStr))
	buf = append(buf, "\x1B[0m ")
	ret := strings.Join(buf, "\x1B[0m ")
	// log.Infof("\x1B[01;37mspeed\x1B[0m 15s/60s/15m \x1B[01;36m%s\x1B[0m \x1B[22;36m%s %s \x1B[01;36mH/s\x1B[0m max: \x1B[01;36m%s H/s\x1B[0m ", fifteenSecondTracker.AverageAsString(), minuteTracker.AverageAsString(), fifteenMinuteTracker.AverageAsString(), "n/a")
	return ret
}

// Add a HashRate entry to all the trackers that are part of this HashRateTrackerArray
func (o HashRateTrackerArray) Add(hr *HashRate) {
	for _, hrt := range o {
		hrt.Add(hr)
	}
}

// SetupHashRateTrackers sets up multiple hashrate trackers using the specified
// inChan as a source of HashRate events. Every duration, the hashrate trackers
// are published to outChan as a HashRateTrackerArray
func SetupHashRateTrackers(duration time.Duration, trackerDurations []time.Duration, inChan <-chan *HashRate, outChan chan<- HashRateTrackerArray) {
	trackers := make(HashRateTrackerArray, len(trackerDurations))
	for idx, duration := range trackerDurations {
		trackers[idx] = NewHashRateTracker(duration)
	}

	var startTime time.Time
	firstHash := true
	for hr := range inChan {
		if firstHash {
			startTime = time.Now()
			firstHash = false
		}
		trackers.Add(hr)

		now := time.Now()
		if now.Sub(startTime) > duration {
			outChan <- trackers
			startTime = now
		}
	}
}

// RunDefaultHashRateTrackers sets up the default hashrate trackers as defined
// by DefaultTrackerDurations and runs an infinite loop listening for hashrate
// events and printing them.
// This function is expected to be run in a goroutine
func RunDefaultHashRateTrackers(inChan <-chan *HashRate) {
	outChan := make(chan HashRateTrackerArray)
	go SetupHashRateTrackers(30*time.Second, DefaultTrackerDurations, inChan, outChan)
	for array := range outChan {
		log.Infof(array.String())
	}
}
