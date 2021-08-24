package loader

import (
	"math"
	"sort"
	"testing"
	"time"
)

func TestBackoffTime(t *testing.T) {
	calc := func(try int, duration time.Duration, factor float32) time.Duration {
		return backoffTime(try, loaderOptions{
			sleepBetweenRetries: duration,
			backoffFactor:       factor,
		})
	}

	timeEquals := func(a time.Duration, b time.Duration) bool {
		diff := time.Duration(math.Abs(float64(a - b)))
		return diff < time.Microsecond
	}

	base := 100 * time.Millisecond

	// test we have no exponential backoff growth when factor is set to 1.0
	tries := 1000
	values := make([]time.Duration, tries)
	for i := 0; i < tries; i++ {
		values[i] = calc(i, base, 1.0)
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	min := values[0]
	max := values[len(values)-1]
	if min != max {
		t.Fatalf(
			"expected a constant value of %v independant of try when backoff factor is 1.0, "+
				"but got a min of %v and a max of %v instead after %d tries",
			base, min, max, tries,
		)
	}

	//
	// Tests on 2.0 factor (100% increase in each try)
	//

	// No increase: 100ms * 2^0 = 100ms
	if result := calc(0, base, 2.0); result != base {
		t.Fatalf("expected %v on first try, but got %v instead", base, result)
	}

	// 100ms * 2^1 = 200ms
	expected := 200 * time.Millisecond
	if result := calc(1, base, 2.0); result != expected {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}

	// 100ms * 2^2 = 400ms
	expected = 400 * time.Millisecond
	if result := calc(2, base, 2.0); result != expected {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}

	// 100ms * 2^3 = 800ms
	expected = 800 * time.Millisecond
	if result := calc(3, base, 2.0); result != expected {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}

	//
	// Tests on 1.1 factor (10% increase in each try)
	//

	// No increase: 100ms * 1.1^0 = 100ms
	if result := calc(0, base, 1.1); result != base {
		t.Fatalf("expected %v on first try, but got %v instead", base, result)
	}

	// 100ms * 1.1^1 = 110ms
	expected = 110 * time.Millisecond
	if result := calc(1, base, 1.1); !timeEquals(expected, result) {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}

	// 100ms * 1.1^2 = 121ms
	expected = 121 * time.Millisecond
	if result := calc(2, base, 1.1); !timeEquals(expected, result) {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}

	// 100ms * 1.1^3 = 133.1ms
	expected = 133100 * time.Microsecond
	if result := calc(3, base, 1.1); !timeEquals(expected, result) {
		t.Fatalf("expected %v on first try, but got %v instead", expected, result)
	}
}
