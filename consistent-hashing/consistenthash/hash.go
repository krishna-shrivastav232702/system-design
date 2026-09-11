package consistenthash

import (
	"crypto/sha256"
	"encoding/binary"
	"hash/crc32"
)


type Hashfunc func(key string) uint32

func defaultHash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key)) //as hashing works on bytes not on strings
}

//alternative

func sha256Hash(key string) uint32{
	sum := sha256.Sum256([]byte(key))
	return binary.BigEndian.Uint32(sum[:4])
}

