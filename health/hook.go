package health

import (
	"sync"

	"github.com/ryanfaerman/netctl/hook"
)

type Check struct {
	Points map[string]error
	mtx    sync.Mutex
}

func (c *Check) Add(name string, err error) {
	c.mtx.Lock()
	c.Points[name] = err
	c.mtx.Unlock()
}

var Hook = hook.New[*Check]("health.check").WithLimit(10)
