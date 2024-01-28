package services

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	circuit "github.com/rubyist/circuitbreaker"
)

type breakerStats struct{}

func (b *breakerStats) Counter(sampleRate float32, bucket string, n ...int) {
	fmt.Println("counter")
	spew.Dump(sampleRate, bucket, n)
}

func (b *breakerStats) Timing(sampleRate float32, bucket string, d ...time.Duration) {
	fmt.Println("timing")
	spew.Dump(sampleRate, bucket, d)
}

func (b *breakerStats) Gauge(sampleRate float32, bucket string, value ...string) {
	fmt.Println("gauge")
	spew.Dump(sampleRate, bucket, value)
}

type breaker struct {
	panel *circuit.Panel
}

var Breaker = &breaker{
	panel: circuit.NewPanel(),
}

func init() {
	Breaker.panel.Statter = &breakerStats{}
	go func() {
		for e := range Breaker.panel.Subscribe() {
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
}

func (s *breaker) Add(name string) { s.panel.Add(name, circuit.NewBreaker()) }

func (s *breaker) AddWithThreshold(name string, n int64) {
	s.panel.Add(name, circuit.NewThresholdBreaker(n))
}

func (s *breaker) AddWithRate(name string, rate float64, min int64) {
	s.panel.Add(name, circuit.NewRateBreaker(rate, min))
}

func (s *breaker) AddWithConsecutive(name string, n int64) {
	s.panel.Add(name, circuit.NewConsecutiveBreaker(n))
}

func (s *breaker) Call(name string, fn func() error) error {
	cb, ok := s.panel.Get(name)
	if !ok {
		s.Add(name)
		cb, _ = s.panel.Get(name)
		global.log.Warn("breaker not found, adding", "name", name)
	}
	return cb.Call(fn, 0)
}

func (s *breaker) Trip(name string) error {
	cb, ok := s.panel.Get(name)
	if !ok {
		return fmt.Errorf("breaker not found: %s", name)
	}
	cb.Trip()
	return nil
}

func (s *breaker) Reset(name string) error {
	cb, ok := s.panel.Get(name)
	if !ok {
		return fmt.Errorf("breaker not found: %s", name)
	}
	cb.Reset()
	return nil
}
