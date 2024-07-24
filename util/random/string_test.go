package random

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringWithRandomizer(t *testing.T) {
	t.Run("SequentialRandomizer", func(t *testing.T) {
		n := 0
		str, err := StringWithRandomizer(AlphanumericRuneSet, 20, func(exclusive int) (int, error) {
			defer func() {
				n++
			}()
			return n % len(AlphanumericRuneSet), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "abcdefghijklmnopqrst", str)
	})

	t.Run("SteppingRandomizer", func(t *testing.T) {
		n := 0
		str, err := StringWithRandomizer(AlphanumericRuneSet, 20, func(exclusive int) (int, error) {
			defer func() {
				n += 2
			}()
			return n % len(AlphanumericRuneSet), nil
		})
		require.NoError(t, err)
		assert.Equal(t, "acegikmoqsuwyACEGIKM", str)
	})
}
