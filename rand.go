// Copyright The ActForGood Authors.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xrand/blob/main/LICENSE.

package xrand

import (
	cRand "crypto/rand"
	"encoding/binary"
	mRand "math/rand"
	"sync"
	"time"
	"unsafe"
)

// defaultJitterFactor is the factor to apply by default on the jitter.
const defaultJitterFactor = 0.2

// globalRand is a global instance of Rand.
var globalRand *mRand.Rand

// init initializes math rand with a secure random seed.
// Is called automatically by go, only once, on this package first import elsewhere.
func init() {
	globalRand = mRand.New(&lockedSource{src: mRand.NewSource(getRandSeed())})
}

// lockedSource allows a random number generator to be used by multiple goroutines
// concurrently. The code is very similar to math/rand.lockedSource, which is
// unfortunately not exposed.
type lockedSource struct {
	sync.Mutex

	src mRand.Source
}

// Int63...
func (ls *lockedSource) Int63() (n int64) {
	ls.Lock()
	n = ls.src.Int63()
	ls.Unlock()

	return
}

// Seed...
func (ls *lockedSource) Seed(seed int64) {
	ls.Lock()
	ls.src.Seed(seed)
	ls.Unlock()
}

// getRandSeed returns a random seed number.
// Uses crypto/rand for that.
// Related discussions upon security:
// 1. https://github.com/golang/go/issues/11871#issuecomment-126350652
// 2. https://stackoverflow.com/a/35208651
// 3. https://stackoverflow.com/a/54491783
func getRandSeed() int64 {
	var b [8]byte
	_, err := cRand.Read(b[:])
	if err == nil {
		// mask off sign bit to ensure positive number
		return int64(binary.LittleEndian.Uint64(b[:]) & (1<<63 - 1))
	}

	// fallback on the common Unix timestamp
	return time.Now().UnixNano()
}

// Intn generates a random integer in range [0,n).
// It panics if max <= 0.
func Intn(n int) int {
	return globalRand.Intn(n)
}

// IntnBetween generates a random integer in range [min,max).
// It panics if max <= 0.
func IntnBetween(min, max int) int {
	return globalRand.Intn(max-min) + min
}

// Float64 generates a random float64 in range [0.0, 1.0).
func Float64() float64 {
	return globalRand.Float64()
}

// Jitter returns a time.Duration altered with a random factor.
// This allows clients to avoid converging on periodic behaviour.
// If maxFactor is <= 0.0, a suggested default value will be chosen.
func Jitter(duration time.Duration, maxFactor ...float64) time.Duration {
	// Note: credits to https://github.com/kubernetes/apimachinery/blob/v0.24.2/pkg/util/wait/wait.go#L196
	factor := defaultJitterFactor
	if len(maxFactor) > 0 && maxFactor[0] > 0.0 {
		factor = maxFactor[0]
	}

	newDuration := time.Duration(0)
	for newDuration <= 0 {
		randRange := 2*Float64() - 1 // [-1.0, 1.0)
		jitter := time.Duration(randRange * factor * float64(duration))
		newDuration = duration + jitter
	}

	return newDuration
}

const (
	// AlphanumAlphabet consists of Ascii lowercase letters, and digits.
	AlphanumAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"
	// DigitsAlphabet consists of 1..9 numbers.
	DigitsAlphabet = "0123456789"
)

// String generates a random string of length n with letters from the alphabet.
// Alphabet is optional and defaults to [AlphanumAlphabet] if not provided.
func String(n int, alphabet ...string) string {
	// Note: implementation details are explained here: https://stackoverflow.com/a/31832326
	// See also similar impl: https://github.com/kubernetes/apimachinery/blob/v0.27.3/pkg/util/rand/rand.go#L98
	var a string
	if len(alphabet) > 0 && len(alphabet[0]) > 0 {
		a = alphabet[0]
	} else {
		a = AlphanumAlphabet
	}

	var (
		alphabetIdxBits       = countBits(len(a))      // represents the max no. of bits to represent an index in alphabet.
		alphabetIdxMask int64 = 1<<alphabetIdxBits - 1 // 1...1b bits, of length alphabetIdxBits
		alphabetIdxMax        = 63 / alphabetIdxBits   // no. of random letters/their indexes we can extract from an int63
		b                     = make([]byte, n)
	)

	randomInt63 := globalRand.Int63()
	remaining := alphabetIdxMax
	for i := 0; i < n; {
		if remaining == 0 { // generate a new random 63 bits integer, reset remaining
			randomInt63, remaining = globalRand.Int63(), alphabetIdxMax
		}
		if alphabetIdx := int(randomInt63 & alphabetIdxMask); alphabetIdx < len(a) {
			b[i] = a[alphabetIdx]
			i++
		}
		randomInt63 >>= alphabetIdxBits
		remaining--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// countBits returns the no. of bits provided integer fits in.
func countBits(x int) int {
	bitsNo := 0
	for x--; x > 0; x >>= 1 {
		bitsNo++
	}

	return bitsNo
}
