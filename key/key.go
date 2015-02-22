package key

import (
	"math"
)

func Decrypt(z uint) (uint, uint) {
	w := uint(math.Floor((math.Sqrt(float64(8*z+1)) - 1) / 2))
	t := (w*w + w) / 2
	y := z - t
	x := w - y
	return x, y
}

func Encrypt(k1, k2 uint) uint {
	return (k1+k2)*(k1+k2+1)/2 + k2
}
