package miner

import (
	"fmt"
	"math"
	"time"

	log "github.com/sirupsen/logrus"
)

type HashRate struct {
	Hashes uint32
	Time   time.Time
}

type hashRateTracker struct {
	hashRates []*HashRate
	duration  time.Duration
	hashes    uint32
}

func NewHashRateTracker(duration time.Duration) *hashRateTracker {
	hrt := &hashRateTracker{}
	hrt.hashRates = make([]*HashRate, 0)
	hrt.duration = duration
	hrt.hashes = 0
	return hrt
}

func (hd *hashRateTracker) Add(hr *HashRate) {
	hd.hashRates = append(hd.hashRates, hr)
	duration := hr.Time.Sub(hd.hashRates[0].Time)
	if duration > hd.duration {
		hd.hashes -= hd.hashRates[0].Hashes
		hd.hashRates = hd.hashRates[1:]
	}
	hd.hashes += hr.Hashes
}

func (hd *hashRateTracker) Average() uint32 {
	duration := hd.hashRates[len(hd.hashRates)-1].Time.Sub(hd.hashRates[0].Time)
	if math.Abs(float64(hd.duration-duration)) > 0.1*float64(hd.duration) {
		return 0
	}
	return uint32(float64(hd.hashes) / duration.Seconds())
}

func (hd *hashRateTracker) AverageAsString() string {
	avg := hd.Average()
	if avg == 0 {
		return "n/a"
	}
	return fmt.Sprintf("%d", avg)
}

func SetupHashRateLogger(hashrateChan chan *HashRate) {
	fifteenSecondTracker := NewHashRateTracker(15 * time.Second)
	minuteTracker := NewHashRateTracker(1 * time.Minute)
	fifteenMinuteTracker := NewHashRateTracker(15 * time.Minute)

	duration := 10 * time.Second

	var startTime time.Time
	firstHash := true
	for hr := range hashrateChan {
		if firstHash {
			startTime = time.Now()
			firstHash = false
		}
		fifteenSecondTracker.Add(hr)
		minuteTracker.Add(hr)
		fifteenMinuteTracker.Add(hr)

		now := time.Now()
		if now.Sub(startTime) > duration {
			log.Infof("\x1B[01;37mspeed\x1B[0m 15s/60s/15m \x1B[01;36m%s\x1B[0m \x1B[22;36m%s %s \x1B[01;36mH/s\x1B[0m max: \x1B[01;36m%s H/s\x1B[0m ", fifteenSecondTracker.AverageAsString(), minuteTracker.AverageAsString(), fifteenMinuteTracker.AverageAsString(), "n/a")
			startTime = now
		}
	}
}
