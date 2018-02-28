package miner

import (
	"time"

	"github.com/fatih/set"
)

type HashRate struct {
	Hashes    uint32
	TimeTaken time.Duration
}

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

func (m *Miner) InformHashrate(hashes uint32, timeTaken time.Duration) {
	data := &HashRate{
		hashes,
		timeTaken,
	}
	for _, obj := range m.hashrateListeners.List() {
		hrChan := obj.(chan *HashRate)
		hrChan <- data
	}
}
