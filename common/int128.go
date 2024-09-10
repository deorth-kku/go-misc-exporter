package common

import (
	"encoding/binary"
	"math"
	"math/big"

	"github.com/anatol/smart.go"
)

const max_uint64 = float64(math.MaxUint64)

func Uint128toFloat64(in smart.Uint128) float64 {
	return float64(in.Val[0]) + max_uint64*float64(in.Val[1])
}

func BigFromInt128(int128 smart.Uint128) *big.Int {
	b := make([]byte, 16)
	binary.BigEndian.PutUint64(b[:8], int128.Val[1])
	binary.BigEndian.PutUint64(b[8:], int128.Val[0])

	result := new(big.Int).SetBytes(b)
	return result
}
