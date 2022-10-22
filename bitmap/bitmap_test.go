package bitmap

import (
	"fmt"
	"testing"
)

func TestBitMap(t *testing.T) {
	bitmap := NewBitMap(24)
	fmt.Printf("%08b\n", bitmap.bits)

	// set
	bitmap.Set(11)
	bitmap.Set(12)
	fmt.Printf("%08b\n", bitmap.bits)

	ok := bitmap.Check(11)
	if !ok {
		t.Fatalf("%08b\n", bitmap.bits)
	}
	ok = bitmap.Check(12)
	if !ok {
		t.Fatalf("%08b\n", bitmap.bits)
	}

	bitmap.Unset(11)
	fmt.Printf("%08b\n", bitmap.bits)
	ok = bitmap.Check(11)
	if ok {
		t.Fatalf("%08b\n", bitmap.bits)
	}
	bitmap.Unset(12)
	fmt.Printf("%08b\n", bitmap.bits)
	ok = bitmap.Check(12)
	if ok {
		t.Fatalf("%08b\n", bitmap.bits)
	}
}
