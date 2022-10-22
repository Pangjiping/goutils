package bloom

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestBloom(t *testing.T) {
	filter := NewBloomFilter(1024)
	filter.Set("aaa")
	filter.Set("bbb")

	t.Log(filter.Check("aaa"))
	t.Log(filter.Check("bbb"))
	t.Log(filter.Check("ccc"))
}

func TestBloom2(t *testing.T) {
	bf := NewBloomFilter()
	mm := make(map[string]bool, 10000)
	for i := 0; i < 10000; i++ {
		val := rand.Int() % 10000
		valStr := fmt.Sprintf("%d", val)
		bf.Set(valStr)
		mm[valStr] = true
	}
	for i := 0; i < 10000; i++ {
		val := rand.Int() % 10000
		valStr := fmt.Sprintf("%d", val)
		ret1 := bf.Check(valStr)
		ret2, _ := mm[valStr]
		if ret1 != ret2 {
			t.Log("error:" + valStr)
		}
	}
}
