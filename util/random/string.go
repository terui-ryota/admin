package random

import (
	"fmt"
)

type RuneSet []rune

var (
	numeric                  = "1234567890"
	lowerAlphabetic          = "abcdefghijklmnopqrstuvwxyz"
	upperAlphabetic          = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	NumericRuneSet           = RuneSet(numeric)
	LowerAlphabeticRuneSet   = RuneSet(lowerAlphabetic)
	LowerAlphanumericRuneSet = RuneSet(lowerAlphabetic + numeric)
	UpperAlphabeticRuneSet   = RuneSet(upperAlphabetic)
	UpperAlphanumericRuneSet = RuneSet(upperAlphabetic + numeric)
	AlphanumericRuneSet      = RuneSet(lowerAlphabetic + upperAlphabetic + numeric)
	LowerHexRuneSet          = RuneSet("abcdef" + numeric)
	UpperHexRuneSet          = RuneSet("ABCDEF" + numeric)
)

// String は 与えられたRuneSetを利用した n 文字のランダムな文字列を生成します
//
// この関数は疑似乱数生成器に math/rand を利用しているため、暗号学的には安全ではありません。
// セキュリティ的な用途には StringWithRandomizer を利用して暗号学的に安全な疑似乱数生成器を利用してください
func String(runes RuneSet, n int) string {
	rnd, release := GetRand()
	defer release()

	str, _ := StringWithRandomizer(runes, n, func(exclusive int) (int, error) {
		return rnd.Intn(exclusive), nil
	})
	return str
}

// StringWithRandomizer は randomizer に与えられた疑似乱数生成器を用いて与えられたRuneSetを利用した n 文字のランダムな文字列を生成します
func StringWithRandomizer(runes RuneSet, n int, randomizer func(exclusive int) (int, error)) (string, error) {
	b := make([]rune, n)
	for i := range b {
		rn, err := randomizer(len(runes))
		if err != nil {
			return "", fmt.Errorf("failed to get random: %w", err)
		}
		b[i] = runes[rn]
	}
	return string(b), nil
}
