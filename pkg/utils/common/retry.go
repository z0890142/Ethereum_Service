package common

import (
	"math"
	"sync/atomic"
	"time"
	"unsafe"
)

type state [2]time.Duration

type fibonacciBackoff struct {
	state unsafe.Pointer
	base  time.Duration
}
type Backoff interface {
	Next() (next time.Duration)
	Reset()
}

func NewFibonacci(base time.Duration) Backoff {
	if base <= 0 {
		panic("base must be greater than 0")
	}

	return &fibonacciBackoff{
		base:  base,
		state: unsafe.Pointer(&state{0, base}),
	}
}

func (b *fibonacciBackoff) Next() time.Duration {
	for {
		curr := atomic.LoadPointer(&b.state)
		currState := (*state)(curr)
		next := currState[0] + currState[1]

		if next <= 0 {
			return math.MaxInt64
		}

		if atomic.CompareAndSwapPointer(&b.state, curr, unsafe.Pointer(&state{currState[1], next})) {
			return next
		}
	}
}

func (b *fibonacciBackoff) Reset() {
	atomic.StorePointer(&b.state, unsafe.Pointer(&state{0, b.base}))
}
