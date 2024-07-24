package random_test

//
//import (
//	"sync"
//	"testing"
//
//	"github.com/stretchr/testify/assert"
//
//	"github.com/ca-media-nantes/libgo/v2/util/random"
//)
//
//func TestRandPool(t *testing.T) {
//	t.Parallel()
//
//	const n = 1e+3
//	wg := sync.WaitGroup{}
//	wg.Add(n)
//	for i := 0; i < n; i++ {
//		go func() {
//			r, f := random.GetRand() // panicにならない
//			assert.NotNil(t, r)
//			f()
//			wg.Done()
//		}()
//	}
//	wg.Wait()
//}
