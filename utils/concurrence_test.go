package utils

import (
	"sync"
	"testing"
)

func TestMaxC(t *testing.T) {
	total := 0
	var lock sync.Mutex
	MaxC(3, 23, func(start, end int) {
		for i := start; i < end; i++ {
			lock.Lock()
			total += i
			lock.Unlock()
		}
	})
	result := 0
	for i := 0; i < 23; i++ {
		result += i
	}
	if result != total {
		t.Fatalf("result [%d] not equal total [%d]", result, total)
	}
}

func TestMaxC2(t *testing.T) {
	total := 0
	var lock sync.Mutex
	MaxC(10, 10, func(start, end int) {
		for i := start; i < end; i++ {
			lock.Lock()
			total += i
			lock.Unlock()
		}
	})
	result := 0
	for i := 0; i < 10; i++ {
		result += i
	}
	if result != total {
		t.Fatalf("result [%d] not equal total [%d]", result, total)
	}
}
