package main

import (
	"errors"
	"fmt"
	"sync"
	"time"

	circuit "github.com/rubyist/circuitbreaker"
)

// Breaker is a circuit breaker, false is closed, true is open
type Breaker struct {
	Name    string
	Open    bool
	LastErr error
}

type Panel struct {
	Breakers map[string]*Breaker
	m        sync.RWMutex
}

func (p *Panel) Register(name string) {
	p.m.Lock()
	p.Breakers[name] = &Breaker{Name: name}
	p.m.Unlock()
}

func (p *Panel) Use(name string, fn func() error) error {
	p.m.Lock()
	defer p.m.Unlock()

	b, ok := p.Breakers[name]
	if !ok {
		return errors.New("breaker not found")
	}
	if b.Open {
		return b.LastErr
	}

	err := fn()
	if err != nil {
		b.Open = true
		b.LastErr = err
	} else {
		b.Open = false
		b.LastErr = nil
	}

	return err
}

func main() {
	panel := circuit.NewPanel()
	panel.Add("foo", circuit.NewThresholdBreaker(10))
	go func() {
		for e := range panel.Subscribe() {
			switch e.Event {
			case circuit.BreakerTripped:
				fmt.Println("breaker tripped", e.Name)
			case circuit.BreakerReset:
				fmt.Println("breaker reset", e.Name)
			case circuit.BreakerFail:
				fmt.Println("breaker fail", e.Name)
			case circuit.BreakerReady:
				fmt.Println("breaker ready", e.Name)
			}
		}
	}()

	for i := 0; i < 100; i++ {
		cb, _ := panel.Get("foo")
		cb.Call(func() error {
			fmt.Println("running")
			if i%2 == 0 {
				return errors.New("oh no")
			}
			return nil
		}, 0)
	}

	time.Sleep(5 * time.Second)
}
