package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFibonacci(t *testing.T) {
	f := NewFibonacci(1 * time.Second)
	assert.NotNil(t, f)
	assert.Equal(t, 1*time.Second, f.Next())
	assert.Equal(t, 2*time.Second, f.Next())
	assert.Equal(t, 3*time.Second, f.Next())

	f.Reset()
	assert.Equal(t, 1*time.Second, f.Next())
}
