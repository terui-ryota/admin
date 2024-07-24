package random

import (
	"math/rand"
	"sync"
	"time"
)

var randPool = sync.Pool{
	New: func() interface{} {
		return rand.New(rand.NewSource(time.Now().UnixNano()))
	},
}

// GetRand gets *rand.Rand from sync.Pool
// call releaser after use to Put *rand.Rand to pool
// ex. defer releaser()
func GetRand() (r *rand.Rand, releaser func()) {
	//nolint:forcetypeassert
	r = randPool.Get().(*rand.Rand)
	releaser = func() { randPool.Put(r) }
	return
}
