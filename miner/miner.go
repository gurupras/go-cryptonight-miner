package miner

import (
	"time"

	"github.com/fatih/set"
	stratum "github.com/gurupras/go-stratum-client"
	log "github.com/sirupsen/logrus"
)

type Miner struct {
	id                uint32
	hashrateListeners set.Interface
}

type Interface interface {
	Id() uint32
	Run() error
	RegisterHashrateListener(chan *HashRate)
}

func New(id uint32) *Miner {
	m := &Miner{
		id,
		set.New(),
	}
	return m
}

func (m *Miner) Id() uint32 {
	return m.id
}

func (m *Miner) RegisterHashrateListener(hrChan chan *HashRate) {
	m.hashrateListeners.Add(hrChan)
}

func (m *Miner) InformHashrate(hashes uint32) {
	data := &HashRate{
		hashes,
		time.Now(),
	}
	for _, obj := range m.hashrateListeners.List() {
		hrChan := obj.(chan *HashRate)
		hrChan <- data
	}
}

func (m *Miner) LogNewWork(sc *stratum.StratumContext, work *stratum.Work) {
	log.Infof("\x1B[01;35mnew job\x1B[0m from \x1B[01;37m%v\x1B[0m diff \x1B[01;37m%d \x1B[0m ", sc.RemoteAddr(), int(work.Difficulty))
}
