// Copyright 2022 Bogdan Constantinescu.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file or at
// https://github.com/actforgood/xrand/LICENSE.

package xrand_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/actforgood/xrand"
)

func TestIntn(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  int
		subject = xrand.Intn
		tests   = []int{1, 11, 23, 104, 256, 999, 15678}
	)

	for _, testData := range tests {
		n := testData // capture range variable
		t.Run(fmt.Sprintf("[0,%d)", n), func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				// act
				result = subject(n)

				// assert
				assertTrue(t, result >= 0)
				assertTrue(t, result < n)
			}
		})
	}
}

func TestIntnBetween(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  int
		subject = xrand.IntnBetween
		tests   = map[int]int{0: 1, 1: 11, 5: 23, 49: 104, 128: 256, 45: 999, 1000: 15678}
	)

	for idx, testData := range tests {
		// capture range variable
		min := idx
		max := testData
		t.Run(fmt.Sprintf("[%d,%d)", min, max), func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				// act
				result = subject(min, max)

				// assert
				assertTrue(t, result >= min)
				assertTrue(t, result < max)
			}
		})
	}
}

func TestFloat64(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  float64
		subject = xrand.Float64
	)

	for i := 0; i < 1000; i++ {
		// act
		result = subject()

		// assert
		assertTrue(t, result >= 0.0)
		assertTrue(t, result < 1.0)
	}
}

func TestJitter(t *testing.T) {
	t.Parallel()

	t.Run("with default max factor", testJitterWithDefaultMaxFactor)
	t.Run("with custom max factor", testJitterWithCustomMaxFactor)
}

func testJitterWithDefaultMaxFactor(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  time.Duration
		subject = xrand.Jitter
		tests   = [...]struct {
			name          string
			inputDuration time.Duration
			expectedMin   time.Duration
			expectedMax   time.Duration
		}{
			{
				name:          "2s",
				inputDuration: 2 * time.Second,
				expectedMin:   time.Duration(1.6 * float64(time.Second)), // jitter = [-0.4s, 0.4s)
				expectedMax:   time.Duration(2.4 * float64(time.Second)),
			},
			{
				name:          "5m",
				inputDuration: 5 * time.Minute,
				expectedMin:   time.Duration(4 * float64(time.Minute)), // jitter = [-1.0m, 1.0m)
				expectedMax:   time.Duration(6 * float64(time.Minute)),
			},
			{
				name:          "100ms",
				inputDuration: 100 * time.Millisecond,
				expectedMin:   time.Duration(80 * float64(time.Millisecond)), // jitter = [-20ms, 20ms)
				expectedMax:   time.Duration(120 * float64(time.Millisecond)),
			},
		}
	)

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			wasDifferent := false
			for i := 0; i < 1000; i++ {
				// act
				result = subject(test.inputDuration)

				// assert
				assertTrue(t, result >= test.expectedMin)
				assertTrue(t, result < test.expectedMax)

				if result != test.inputDuration {
					wasDifferent = true
				}
			}
			assertTrue(t, wasDifferent)
		})
	}
}

func testJitterWithCustomMaxFactor(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  time.Duration
		subject = xrand.Jitter
		tests   = [...]struct {
			name           string
			inputDuration  time.Duration
			inputMaxFactor float64
			expectedMin    time.Duration
			expectedMax    time.Duration
		}{
			{
				name:           "2s",
				inputDuration:  2 * time.Second,
				inputMaxFactor: 1.0,
				expectedMin:    1, // jitter = [-2s, 2s)
				expectedMax:    time.Duration(4 * float64(time.Second)),
			},
			{
				name:           "5m",
				inputDuration:  5 * time.Minute,
				inputMaxFactor: 1.0,
				expectedMin:    1, // jitter = [-5m, 5m)
				expectedMax:    time.Duration(10 * float64(time.Minute)),
			},
			{
				name:           "100ms",
				inputDuration:  100 * time.Millisecond,
				inputMaxFactor: 1.0,
				expectedMin:    1, // jitter = [-100ms, 100ms)
				expectedMax:    time.Duration(200 * float64(time.Millisecond)),
			},
			{
				name:           "2s",
				inputDuration:  2 * time.Second,
				inputMaxFactor: 0.6,
				expectedMin:    time.Duration(0.8 * float64(time.Second)), // jitter = [-1.2s, 1.2s)
				expectedMax:    time.Duration(3.2 * float64(time.Second)),
			},
			{
				name:           "5m",
				inputDuration:  5 * time.Minute,
				inputMaxFactor: 0.6,
				expectedMin:    time.Duration(2 * float64(time.Minute)), // jitter = [-3m, 3m)
				expectedMax:    time.Duration(8 * float64(time.Minute)),
			},
			{
				name:           "100ms",
				inputDuration:  100 * time.Millisecond,
				inputMaxFactor: 0.6,
				expectedMin:    time.Duration(40 * float64(time.Millisecond)), // jitter = [-60ms, 60ms)
				expectedMax:    time.Duration(160 * float64(time.Millisecond)),
			},
			{
				name:           "positive",
				inputDuration:  1 * time.Nanosecond,
				inputMaxFactor: 100.0,
				expectedMin:    1, // jitter = [-100ns, 100ns)
				expectedMax:    time.Duration(101 * float64(time.Nanosecond)),
			},
		}
	)

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			wasDifferent := false
			for i := 0; i < 1000; i++ {
				// act
				result = subject(test.inputDuration, test.inputMaxFactor)

				// assert
				assertTrue(t, result >= test.expectedMin)
				assertTrue(t, result < test.expectedMax)

				if result != test.inputDuration {
					wasDifferent = true
				}
			}
			assertTrue(t, wasDifferent)
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()

	// arrange
	var (
		result  string
		subject = xrand.String
		tests   = [...]struct {
			name          string
			inputLength   int
			inputAlphabet string
			expectedReg   *regexp.Regexp
		}{
			{
				name:          "len = 10, alphabet = xrand.AlphanumAlphabet",
				inputLength:   10,
				inputAlphabet: xrand.AlphanumAlphabet,
				expectedReg:   regexp.MustCompile(`^[a-z0-9]{10}$`),
			},
			{
				name:          "len = 32, alphabet = xrand.AlphanumAlphabet",
				inputLength:   32,
				inputAlphabet: xrand.AlphanumAlphabet,
				expectedReg:   regexp.MustCompile(`^[a-z0-9]{32}$`),
			},
			{
				name:          "len = 150, alphabet = xrand.AlphanumAlphabet",
				inputLength:   150,
				inputAlphabet: xrand.AlphanumAlphabet,
				expectedReg:   regexp.MustCompile(`^[a-z0-9]{150}$`),
			},
			{
				name:          "len = 43, alphabet = xrand.DigitsAlphabet",
				inputLength:   43,
				inputAlphabet: xrand.DigitsAlphabet,
				expectedReg:   regexp.MustCompile(`^[0-9]{43}$`),
			},
			{
				name:          "len = 109, alphabet = abc-",
				inputLength:   109,
				inputAlphabet: "abc-",
				expectedReg:   regexp.MustCompile(`^[a-c\-]{109}$`),
			},
			{
				name:          "len = 0",
				inputLength:   0,
				inputAlphabet: "abc",
				expectedReg:   regexp.MustCompile(`^$`),
			},
			{
				name:          "empty alphabet - default alphabet",
				inputLength:   2,
				inputAlphabet: "",
				expectedReg:   regexp.MustCompile(`^[a-z0-9]{2}$`),
			},
		}
	)

	for _, testData := range tests {
		test := testData // capture range variable
		t.Run(test.name, func(t *testing.T) {
			for i := 0; i < 1000; i++ {
				// act
				result = subject(test.inputLength, test.inputAlphabet)

				// assert
				assertTrue(t, test.expectedReg.Match([]byte(result)))
			}
		})
	}
}

func BenchmarkString(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = xrand.String(16)
	}
}

func BenchmarkJitter(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = xrand.Jitter(5 * time.Minute)
	}
}

// assertTrue checks if value passed is true.
// Returns successful assertion status.
func assertTrue(t *testing.T, actual bool) bool {
	t.Helper()
	if !actual {
		t.Error("should be true")

		return false
	}

	return true
}

func ExampleIntn() {
	// generate a random int in [0, 1000)
	randInt := xrand.Intn(1000)
	fmt.Println(randInt)
}

func ExampleIntnBetween() {
	// generate a random int in [100, 200)
	randInt := xrand.IntnBetween(100, 200)
	fmt.Println(randInt)
}

func ExampleFloat64() {
	// generate a random float in [0.0, 1.0)
	randFloat := xrand.Float64()
	fmt.Println(randFloat)
}

func ExampleJitter() {
	// slightly alter +/- a time.Duration
	cacheTTL := 10 * time.Minute
	factor := 0.1
	jitteredCacheTTL := xrand.Jitter(cacheTTL, factor)
	fmt.Println(jitteredCacheTTL)
}

func ExampleString() {
	// generate a random string of length 16, containing [a-z0-9] letters.
	randString := xrand.String(16, xrand.AlphanumAlphabet)
	fmt.Println(randString)
}
