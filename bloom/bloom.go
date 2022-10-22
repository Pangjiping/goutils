package bloom

import "github.com/Pangjiping/goutils/bitmap"

type BloomFilter struct {
	bset *bitmap.BitMap
	size uint
}

const defaultBloomFilterSize uint = 1024 * 1024

var seeds = []uint{3011, 3017, 3031}

func NewBloomFilter(size ...uint) *BloomFilter {
	var sizeVal uint
	if len(size) > 0 && size[0] > 0 {
		sizeVal = size[0]
	} else {
		sizeVal = defaultBloomFilterSize
	}

	filter := &BloomFilter{}
	filter.bset = bitmap.NewBitMap(sizeVal)
	filter.size = sizeVal
	return filter
}

func (filter *BloomFilter) hashFunc(seed uint, value string) uint64 {
	hash := uint64(seed)
	for i := 0; i < len(value); i++ {
		hash = hash*33 + uint64(value[i])
	}
	return hash
}

func (filter *BloomFilter) Set(value string) {
	for _, seed := range seeds {
		hash := filter.hashFunc(seed, value)
		hash = hash % uint64(filter.size)
		filter.bset.Set(uint(hash))
	}
}

func (filter *BloomFilter) Check(value string) bool {
	for _, seed := range seeds {
		hash := filter.hashFunc(seed, value)
		hash = hash % uint64(filter.size)
		ret := filter.bset.Check(uint(hash))
		if !ret {
			return false
		}
	}
	return true
}
