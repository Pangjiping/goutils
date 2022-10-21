package radix

import (
	"bytes"
	"fmt"
)

type stringKV struct {
	key string
	val interface{}
}

type raxNode struct {
	bitLen int
	bitVal []byte
	left   *raxNode
	right  *raxNode
	val    interface{}
}

func newRaxNode() *raxNode {
	return &raxNode{
		bitLen: 0,
		bitVal: nil,
		left:   nil,
		right:  nil,
		val:    nil,
	}
}

func (this *raxNode) pathSplit(key []byte, key_pos int, val interface{}) (int, *raxNode) {
	//与path对应的key数据(去掉已经处理的公共字节)
	data := key[key_pos/8:]
	//key以bit为单位长度（包含开始字节的字节内bit位的偏移量）
	bit_end_key := len(data) * 8
	//path以bit为单位长度（包含开始字节的字节内bit位的偏移量）
	bit_end_path := key_pos%8 + int(this.bitLen)
	//当前的bit偏移量，需要越过开始字节的字节内bit位的偏移量
	bpos := key_pos % 8
	for bpos < bit_end_key && bpos < bit_end_path {
		ci := bpos / 8
		byte_path := this.bitVal[ci]
		byte_data := data[ci]

		//起始字节的内部偏移量
		beg := 0
		if ci == 0 {
			beg = key_pos % 8
		}
		//终止字节的内部偏移量
		end := 8
		if ci == bit_end_path/8 {
			end = bit_end_path % 8
			if end == 0 {
				end = 8
			}
		}

		if beg != 0 || end != 8 {
			//不完整字节的比较，若不等则跳出循环
			//若相等，则增长bpos，并继续比较下一个字节
			num := GetPrefixBitLength2(byte_data, byte_path, beg, end)
			bpos += num
			if num < end-beg {
				break
			}
		} else if byte_data != byte_path {
			//完整字节比较，num为相同的bit的个数，bpos增加num后跳出循环
			num := GetPrefixBitLength2(byte_data, byte_path, 0, 8)
			bpos += num
			break
		} else {
			//完整字节比较，相等，则继续比较下一个字节
			bpos += 8
		}
	}

	//当前字节的位置
	char_index := bpos / 8
	//当前字节的bit偏移量
	bit_offset := bpos % 8
	//剩余的path长度
	bit_last_path := bit_end_path - bpos
	//剩余的key长度
	bit_last_data := bit_end_key - bpos

	//key的数据有剩余
	//若path有子结点，则继续处理子结点
	//若path没有子结点，则创建一个key子结点
	var nd_data *raxNode = nil
	var bval_data byte
	if bit_last_data > 0 {
		//若path有子结点，则退出本函数，并在子结点中进行处理
		byte_data := data[char_index]
		bval_data = GET_BIT(byte_data, bit_offset)
		if bit_last_path == 0 {
			if bval_data == 0 && this.left != nil {
				return key_pos + int(this.bitLen), this.left
			} else if bval_data == 1 && this.right != nil {
				return key_pos + int(this.bitLen), this.right
			}
		}

		//为剩余的key创建子结点
		nd_data = newRaxNode()
		nd_data.left = nil
		nd_data.right = nil
		nd_data.val = val
		nd_data.bitLen = bit_last_data
		nd_data.bitVal = make([]byte, len(data[char_index:]))
		copy(nd_data.bitVal, data[char_index:])

		//若bit_offset!=0，说明不是完整字节，
		//将字节分裂，并将字节中的非公共部分分离出来,保存到子结点中
		if bit_offset != 0 {
			byte_tmp := CLEAR_BITS_LOW(byte_data, bit_offset)
			nd_data.bitVal[0] = byte_tmp
		}
	}

	//path的数据有剩余
	//创建子节点：nd_path结点
	//并将数据分开，公共部分保存this结点，其他保存到nd_path结点
	var nd_path *raxNode = nil
	var bval_path byte
	if bit_last_path > 0 {
		byte_path := this.bitVal[char_index]
		bval_path = GET_BIT(byte_path, bit_offset)

		//为剩余的path创建子结点
		nd_path = newRaxNode()
		nd_path.left = this.left
		nd_path.right = this.right
		nd_path.val = this.val
		nd_path.bitLen = bit_last_path
		nd_path.bitVal = make([]byte, len(this.bitVal[char_index:]))
		copy(nd_path.bitVal, this.bitVal[char_index:])

		//将byte_path字节中的非公共部分分离出来,保存到子结点中
		if bit_offset != 0 {
			byte_tmp := CLEAR_BITS_LOW(byte_path, bit_offset)
			nd_path.bitVal[0] = byte_tmp
		}

		//修改当前结点，作为nd_path结点、nd_data结点的父结点
		//多申请一个子节，用于存储可能出现的不完整字节
		bit_val_old := this.bitVal
		this.left = nil
		this.right = nil
		this.val = nil
		this.bitLen = this.bitLen - bit_last_path //=bpos - (key_pos % 8)
		this.bitVal = make([]byte, len(bit_val_old[0:char_index])+1)
		copy(this.bitVal, bit_val_old[0:char_index])
		this.bitVal = this.bitVal[0 : len(this.bitVal)-1]

		//将byte_path字节中的公共部分分离出来,保存到父结点
		if bit_offset != 0 {
			byte_tmp := CLEAR_BITS_HIGH(byte_path, 8-bit_offset)
			this.bitVal = append(this.bitVal, byte_tmp)
		}
	}

	//若path包含key，则将val赋值给this结点
	if bit_last_data == 0 {
		this.val = val
	}
	if nd_data != nil {
		if bval_data == 0 {
			this.left = nd_data
		} else {
			this.right = nd_data
		}
	}
	if nd_path != nil {
		if bval_path == 0 {
			this.left = nd_path
		} else {
			this.right = nd_path
		}
	}
	return len(key) * 8, nil
}

func (this *raxNode) pathCompare(data []byte, bbeg int) (bool, int) {
	bend := bbeg + int(this.bitLen)
	if bend > len(data)*8 {
		return false, len(data) * 8
	}

	//起始和终止字节的位置
	cbeg := bbeg / 8
	cend := bend / 8
	//起始和终止字节的偏移量
	obeg := bbeg % 8
	oend := bend % 8
	for bb := bbeg; bb < bend; {
		//获取两个数组的当前字节位置
		dci := bb / 8
		nci := dci - cbeg

		//获取数据的当前字节以及循环步长
		step := 8
		byte_data := data[dci]
		if dci == cbeg && obeg > 0 {
			//清零不完整字节的低位
			byte_data = CLEAR_BITS_LOW(byte_data, obeg)
			step -= obeg
		}
		if dci == cend && oend > 0 {
			//清零不完整字节的高位
			byte_data = CLEAR_BITS_HIGH(byte_data, 8-oend)
			step -= 8 - oend
		}

		//获取结点的当前字节，并与数据的当前字节比较
		byte_node := this.bitVal[nci]
		if byte_data != byte_node {
			return false, len(data) * 8
		}

		bb += step
	}

	return true, bend
}

//将当前结点的子结点进行合并
//若当前结点只有一个子结点，并且当前结点是空结点，才可以进行合并操作
func (this *raxNode) pathMerge(bpos int) bool {
	//若当前结点存在值，则不能合并
	if this.val != nil {
		return false
	}

	//若当前结点有2个子结点，则不能合并
	if this.left != nil && this.right != nil {
		return false
	}

	//若当前结点没有子结点，则不能合并
	if this.left != nil && this.right != nil {
		return false
	}

	//获取当前结点的子结点
	child := this.left
	if this.right != nil {
		child = this.right
	}

	//判断当前结点最后一个字节是否是完整的字节
	//若不是完整字节，需要与子结点的第一个字节进行合并
	if bpos%8 != 0 {
		char_len := len(this.bitVal)
		char_last := this.bitVal[char_len-1]
		char_0000 := child.bitVal[0]
		child.bitVal = child.bitVal[1:]
		this.bitVal[char_len-1] = char_last | char_0000
	}

	//合并当前结点以及子结点
	this.val = child.val
	this.bitVal = append(this.bitVal, child.bitVal...)
	this.bitLen += child.bitLen
	this.left = child.left
	this.right = child.right

	return true
}

////打印结点信息，用于调试
func (this *raxNode) GetNodeInfo(bbeg int) string {
	buff := new(bytes.Buffer)

	bend := bbeg + int(this.bitLen)
	//起始和终止字节的位置
	cbeg := bbeg / 8
	cend := bend / 8
	//起始和终止字节的偏移量
	obeg := bbeg % 8
	oend := bend % 8
	for bb := bbeg; bb < bend; {
		//获取两个数组的当前字节位置
		dci := bb / 8
		nci := dci - cbeg
		byte_node := this.bitVal[nci]

		//获取数据的当前字节以及循环步长
		step := 8
		if nci == 0 && obeg > 0 {
			step = 8 - obeg
		}
		if dci == cend && oend > 0 {
			step = oend
		}
		if cbeg == cend {
			step = int(this.bitLen)
		}

		if step != 8 {
			buff.WriteString(fmt.Sprintf("(%08b:%d)", byte_node, byte_node))
		} else {
			buff.WriteByte(byte_node)
		}
		bb += step
	}

	if this.val != nil {
		buff.WriteString(fmt.Sprintf("=%v", this.val))
	}

	return buff.String()
}
