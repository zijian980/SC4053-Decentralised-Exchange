package nonce

import (
	"github.com/ethereum/go-ethereum/common"
	"sync"
)

type NonceRegistry struct {
	Registry map[common.Address]*Counter
}

type Counter struct {
	Nonce int
	mu    sync.Mutex
}

func NewNonceRegistry() *NonceRegistry {
	return &NonceRegistry{
		Registry: make(map[common.Address]*Counter),
	}
}

func (r *NonceRegistry) get(addr common.Address) *Counter {
	val, ok := r.Registry[addr]
	if ok {
		return val
	}
	return nil
}

func (r *NonceRegistry) add(addr common.Address) *Counter {
	newCounter := &Counter{
		Nonce: 0,
		mu:    sync.Mutex{},
	}
	r.Registry[addr] = newCounter
	return newCounter
}

func (r *NonceRegistry) Inc(addr common.Address) int {
	counter := r.get(addr)
	if counter == nil {
		counter = r.add(addr)
	}
	counter.Inc()
	return counter.Nonce
}

func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Nonce++
}
