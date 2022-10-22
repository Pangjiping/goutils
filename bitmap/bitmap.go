package bitmap

type BitMap struct {
	bits []byte
	vmax uint
}

const defaultVMax uint = 8192

func NewBitMap(maxVal ...uint) *BitMap {
	var vmax uint
	if len(maxVal) > 0 && maxVal[0] > 0 {
		vmax = maxVal[0]
	} else {
		vmax = defaultVMax
	}

	bm := &BitMap{}
	bm.vmax = vmax
	size := (vmax + 7) / 8
	bm.bits = make([]byte, size, size)
	return bm
}

func (m *BitMap) Set(num uint) {
	if num > m.vmax {
		m.vmax += 1024
		if m.vmax < num {
			m.vmax = num
		}

		dd := int(num+7)/8 - len(m.bits)
		if dd > 0 {
			tempArr := make([]byte, dd, dd)
			m.bits = append(m.bits, tempArr...)
		}
	}

	m.bits[num/8] |= 1 << (num % 8)
}

func (m *BitMap) Unset(num uint) {
	if num > m.vmax {
		return
	}
	m.bits[num/8] &^= 1 << (num % 8)
}

func (m *BitMap) Check(num uint) bool {
	if num > m.vmax {
		return false
	}
	return m.bits[num/8]&(1<<(num%8)) != 0
}
