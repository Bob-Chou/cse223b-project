package hash

import (
	"hash/fnv"
)

const (
	// MaxHashBits is the max bit of hash result
	MaxHashBits int = 16
)

// EncodeKey takes a key and hash to uint64
func EncodeKey(str string) uint64 {
	h := fnv.New64()
	_, _ = h.Write([]byte("dtCI4aKoFG" + str + "yaeAphu2QR"))
	return h.Sum64() % (1 << MaxHashBits)
}
