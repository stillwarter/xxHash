// Package xxHash32 implements the very fast xxHash hashing algorithm (32 bits version).
// (https://github.com/Cyan4973/xxHash/)
package xxHash32

import (
	"hash"
	"unsafe"
)

const (
	prime32_1 = 2654435761
	prime32_2 = 2246822519
	prime32_3 = 3266489917
	prime32_4 = 668265263
	prime32_5 = 374761393
)

var littleEndian bool

func init() {
	b := [4]byte{1, 2, 3, 4}
	actual := *(*uint32)(unsafe.Pointer(&b[0]))
	expected := uint32(0)
	for i, v := range b {
		expected += uint32(v) << (uint(i) * 8)
	}
	littleEndian = actual == expected
}

type xxHash struct {
	seed     uint32
	v1       uint32
	v2       uint32
	v3       uint32
	v4       uint32
	totalLen uint64
	buf      [16]byte
	bufused  int
}

// New returns a new Hash32 instance.
func New(seed uint32) hash.Hash32 {
	xxh := &xxHash{seed: seed}
	xxh.Reset()
	return xxh
}

// Sum appends the current hash to b and returns the resulting slice.
// It does not change the underlying hash state.
func (xxh xxHash) Sum(b []byte) []byte {
	h32 := xxh.Sum32()
	return append(b, byte(h32), byte(h32>>8), byte(h32>>16), byte(h32>>24))
}

// Reset resets the Hash to its initial state.
func (xxh *xxHash) Reset() {
	xxh.v1 = xxh.seed + prime32_1 + prime32_2
	xxh.v2 = xxh.seed + prime32_2
	xxh.v3 = xxh.seed
	xxh.v4 = xxh.seed - prime32_1
	xxh.totalLen = 0
	xxh.bufused = 0
}

// Size returns the number of bytes returned by Sum().
func (xxh *xxHash) Size() int {
	return 4
}

// BlockSize gives the minimum number of bytes accepted by Write().
func (xxh *xxHash) BlockSize() int {
	return 1
}

// Write adds input bytes to the Hash.
// It never returns an error.
func (xxh *xxHash) Write(input []byte) (int, error) {
	n := len(input)
	m := xxh.bufused

	xxh.totalLen += uint64(n)

	r := len(xxh.buf) - m
	if n < r {
		copy(xxh.buf[m:], input)
		xxh.bufused += len(input)
		return n, nil
	}

	p := 0
	if m > 0 {
		// some data left from previous update
		copy(xxh.buf[xxh.bufused:], input[:r])
		xxh.bufused += len(input) - r

		// fast rotl(13)
		if littleEndian {
			ptr := uintptr(unsafe.Pointer(&xxh.buf[p]))
			p32 := xxh.v1 + *(*uint32)(unsafe.Pointer(ptr))*prime32_2
			xxh.v1 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v2 + *(*uint32)(unsafe.Pointer(ptr + 4))*prime32_2
			xxh.v2 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v3 + *(*uint32)(unsafe.Pointer(ptr + 8))*prime32_2
			xxh.v3 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v4 + *(*uint32)(unsafe.Pointer(ptr + 12))*prime32_2
			xxh.v4 = (p32<<13 | p32>>19) * prime32_1
		} else {
			p32 := xxh.v1 + (uint32(xxh.buf[0+3])<<24|uint32(xxh.buf[0+2])<<16|uint32(xxh.buf[0+1])<<8|uint32(xxh.buf[0]))*prime32_2
			xxh.v1 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v2 + (uint32(xxh.buf[4+3])<<24|uint32(xxh.buf[4+2])<<16|uint32(xxh.buf[4+1])<<8|uint32(xxh.buf[4]))*prime32_2
			xxh.v2 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v3 + (uint32(xxh.buf[8+3])<<24|uint32(xxh.buf[8+2])<<16|uint32(xxh.buf[8+1])<<8|uint32(xxh.buf[8]))*prime32_2
			xxh.v3 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v4 + (uint32(xxh.buf[12+3])<<24|uint32(xxh.buf[12+2])<<16|uint32(xxh.buf[12+1])<<8|uint32(xxh.buf[12]))*prime32_2
			xxh.v4 = (p32<<13 | p32>>19) * prime32_1
		}
		p = r
		xxh.bufused = 0
	}

	if p > n-16 {
		// Nothing to do
	} else if littleEndian {
		ptr := uintptr(unsafe.Pointer(&input[p]))
		for ; p <= n-16; p += 16 {
			p32 := xxh.v1 + *(*uint32)(unsafe.Pointer(ptr))*prime32_2
			xxh.v1 = (p32<<13 | p32>>19) * prime32_1
			ptr += 4
			p32 = xxh.v2 + *(*uint32)(unsafe.Pointer(ptr))*prime32_2
			xxh.v2 = (p32<<13 | p32>>19) * prime32_1
			ptr += 4
			p32 = xxh.v3 + *(*uint32)(unsafe.Pointer(ptr))*prime32_2
			xxh.v3 = (p32<<13 | p32>>19) * prime32_1
			ptr += 4
			p32 = xxh.v4 + *(*uint32)(unsafe.Pointer(ptr))*prime32_2
			xxh.v4 = (p32<<13 | p32>>19) * prime32_1
			ptr += 4
		}
	} else {
		for n := n - 16; p <= n; p += 16 {
			sub := input[p : p+16]
			p32 := xxh.v1 + (uint32(sub[0+3])<<24|uint32(sub[0+2])<<16|uint32(sub[0+1])<<8|uint32(sub[0]))*prime32_2
			xxh.v1 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v2 + (uint32(sub[4+3])<<24|uint32(sub[4+2])<<16|uint32(sub[4+1])<<8|uint32(sub[4]))*prime32_2
			xxh.v2 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v3 + (uint32(sub[8+3])<<24|uint32(sub[8+2])<<16|uint32(sub[8+1])<<8|uint32(sub[8]))*prime32_2
			xxh.v3 = (p32<<13 | p32>>19) * prime32_1
			p32 = xxh.v4 + (uint32(sub[12+3])<<24|uint32(sub[12+2])<<16|uint32(sub[12+1])<<8|uint32(sub[12]))*prime32_2
			xxh.v4 = (p32<<13 | p32>>19) * prime32_1
		}
	}

	copy(xxh.buf[xxh.bufused:], input[p:])
	xxh.bufused += len(input) - p

	return n, nil
}

// Sum32 returns the 32 bits Hash value.
func (xxh *xxHash) Sum32() uint32 {
	h32 := uint32(xxh.totalLen)
	if xxh.totalLen >= 16 {
		h32 += ((xxh.v1 << 1) | (xxh.v1 >> 31)) +
			((xxh.v2 << 7) | (xxh.v2 >> 25)) +
			((xxh.v3 << 12) | (xxh.v3 >> 20)) +
			((xxh.v4 << 18) | (xxh.v4 >> 14))
	} else {
		h32 += xxh.seed + prime32_5
	}

	if littleEndian {
		p := 0
		n := xxh.bufused
		ptr := uintptr(unsafe.Pointer(&xxh.buf[0]))
		for n := n - 4; p <= n; p += 4 {
			h32 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_3
			h32 = ((h32 << 17) | (h32 >> 15)) * prime32_4
			ptr += 4
		}
		for ; p < n; p++ {
			h32 += uint32(xxh.buf[p]) * prime32_5
			h32 = ((h32 << 11) | (h32 >> 21)) * prime32_1
		}
	} else {
		p := 0
		n := xxh.bufused
		for n := n - 4; p <= n; p += 4 {
			h32 += (uint32(xxh.buf[p+3])<<24 | uint32(xxh.buf[p+2])<<16 | uint32(xxh.buf[p+1])<<8 | uint32(xxh.buf[p])) * prime32_3
			h32 = ((h32 << 17) | (h32 >> 15)) * prime32_4
		}
		for ; p < n; p++ {
			h32 += uint32(xxh.buf[p]) * prime32_5
			h32 = ((h32 << 11) | (h32 >> 21)) * prime32_1
		}
	}

	h32 ^= h32 >> 15
	h32 *= prime32_2
	h32 ^= h32 >> 13
	h32 *= prime32_3
	h32 ^= h32 >> 16

	return h32
}

// Checksum returns the 32bits Hash value.
func Checksum(input []byte, seed uint32) uint32 {
	n := len(input)
	h32 := uint32(n)

	if n < 16 {
		h32 += seed + prime32_5
	} else {
		v1 := seed + prime32_1 + prime32_2
		v2 := seed + prime32_2
		v3 := seed
		v4 := seed - prime32_1
		p := 0
		if n < 16 {
			// Nothing to do
		} else if littleEndian {
			ptr := uintptr(unsafe.Pointer(&input[p]))
			for n := n - 16; p <= n; p += 16 {
				v1 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_2
				v1 = (v1<<13 | v1>>19) * prime32_1
				ptr += 4
				v2 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_2
				v2 = (v2<<13 | v2>>19) * prime32_1
				ptr += 4
				v3 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_2
				v3 = (v3<<13 | v3>>19) * prime32_1
				ptr += 4
				v4 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_2
				v4 = (v4<<13 | v4>>19) * prime32_1
				ptr += 4
			}
		} else {
			for n := n - 16; p <= n; p += 16 {
				sub := input[p : p+16]
				v1 += (uint32(sub[3])<<24 | uint32(sub[2])<<16 | uint32(sub[1])<<8 | uint32(sub[0])) * prime32_2
				v1 = (v1<<13 | v1>>19) * prime32_1
				v2 += (uint32(sub[4+3])<<24 | uint32(sub[4+2])<<16 | uint32(sub[4+1])<<8 | uint32(sub[4])) * prime32_2
				v2 = (v2<<13 | v2>>19) * prime32_1
				v3 += (uint32(sub[8+3])<<24 | uint32(sub[8+2])<<16 | uint32(sub[8+1])<<8 | uint32(sub[8])) * prime32_2
				v3 = (v3<<13 | v3>>19) * prime32_1
				v4 += (uint32(sub[12+3])<<24 | uint32(sub[12+2])<<16 | uint32(sub[12+1])<<8 | uint32(sub[12])) * prime32_2
				v4 = (v4<<13 | v4>>19) * prime32_1
			}
		}
		input = input[p:]
		n -= p
		h32 += ((v1 << 1) | (v1 >> 31)) +
			((v2 << 7) | (v2 >> 25)) +
			((v3 << 12) | (v3 >> 20)) +
			((v4 << 18) | (v4 >> 14))
	}

	if n == 0 {
		// Nothing to do
	} else if littleEndian {
		p := 0
		ptr := uintptr(unsafe.Pointer(&input[0]))
		for n := n - 4; p <= n; p += 4 {
			h32 += *(*uint32)(unsafe.Pointer(ptr)) * prime32_3
			h32 = ((h32 << 17) | (h32 >> 15)) * prime32_4
			ptr += 4
		}
		for p < n {
			h32 += uint32(input[p]) * prime32_5
			h32 = ((h32 << 11) | (h32 >> 21)) * prime32_1
			p++
		}
	} else {
		p := 0
		for n := n - 4; p <= n; p += 4 {
			sub := input[p : p+4]
			h32 += (uint32(sub[3])<<24 | uint32(sub[2])<<16 | uint32(sub[1])<<8 | uint32(sub[0])) * prime32_3
			h32 = ((h32 << 17) | (h32 >> 15)) * prime32_4
		}
		for p < n {
			h32 += uint32(input[p]) * prime32_5
			h32 = ((h32 << 11) | (h32 >> 21)) * prime32_1
			p++
		}
	}

	h32 ^= h32 >> 15
	h32 *= prime32_2
	h32 ^= h32 >> 13
	h32 *= prime32_3
	h32 ^= h32 >> 16

	return h32
}
