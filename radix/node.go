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

// pathSplit performs path splitting when inserting data.
func (radixNode *raxNode) pathSplit(key []byte, keyPos int, val interface{}) (int, *raxNode) {
	// the key data corresponding to the path
	// (remove the processed common bytes)
	data := key[keyPos/8:]

	// the length of the key in bits
	// (the offset of the bit in the byte containing the start byte)
	bitEndKey := len(data) * 8

	// path length in bits
	// (bit offset within the byte containing the start byte)
	bitEndPath := keyPos%8 + int(radixNode.bitLen)

	// the current bit offset, the offset of the bit in the byte that needs to go beyond the start byte
	bpos := keyPos % 8
	for bpos < bitEndKey && bpos < bitEndPath {
		ci := bpos / 8
		bytePath := radixNode.bitVal[ci]
		byteData := data[ci]

		// internal offset of start byte
		beg := 0
		if ci == 0 {
			beg = keyPos % 8
		}
		// internal offset of the termination byte
		end := 8
		if ci == bitEndPath/8 {
			end = bitEndPath % 8
			if end == 0 {
				end = 8
			}
		}

		if beg != 0 || end != 8 {
			// comparison of incomplete bytes, if not equal, jump out of the loop;
			// if equal, increment bpos and continue to compare the next byte.
			num := GetPrefixBitLength2(byteData, bytePath, beg, end)
			bpos += num
			if num < end-beg {
				break
			}
		} else if byteData != bytePath {
			// complete byte comparison, num is the same number of bits, bpos increases num and jumps out of the loop.
			num := GetPrefixBitLength2(byteData, bytePath, 0, 8)
			bpos += num
			break
		} else {
			// complete byte comparison, if equal, continue to compare the next byte
			bpos += 8
		}
	}

	// current byte position
	charIndex := bpos / 8
	// bit offset of the current byte
	bitOffset := bpos % 8
	// remaining path length
	bitLastPath := bitEndPath - bpos
	// remaining key length
	bitLastData := bitEndKey - bpos

	// the data of the key is left over.
	// If path has child nodes, continue processing child nodes;
	// If path has no child nodes, create a key child node.
	var nd_data *raxNode = nil
	var bval_data byte
	if bitLastData > 0 {
		// If path has child nodes, exit this function and process in child nodes
		byte_data := data[charIndex]
		bval_data = GET_BIT(byte_data, bitOffset)
		if bitLastPath == 0 {
			if bval_data == 0 && radixNode.left != nil {
				return keyPos + int(radixNode.bitLen), radixNode.left
			} else if bval_data == 1 && radixNode.right != nil {
				return keyPos + int(radixNode.bitLen), radixNode.right
			}
		}

		// Create child nodes for the remaining keys
		nd_data = newRaxNode()
		nd_data.left = nil
		nd_data.right = nil
		nd_data.val = val
		nd_data.bitLen = bitLastData
		nd_data.bitVal = make([]byte, len(data[charIndex:]))
		copy(nd_data.bitVal, data[charIndex:])

		// If bit_offset!=0, it means not a complete byte
		// Split the bytes, separate the non-common parts of the bytes, and save them to the child nodes
		if bitOffset != 0 {
			byte_tmp := CLEAR_BITS_LOW(byte_data, bitOffset)
			nd_data.bitVal[0] = byte_tmp
		}
	}

	// The data of path has remaining
	// Create child nodes: nd_path node
	// And separate the data, save the common part to the radixNode node, and save the other to the nd_path node
	var nd_path *raxNode = nil
	var bval_path byte
	if bitLastPath > 0 {
		byte_path := radixNode.bitVal[charIndex]
		bval_path = GET_BIT(byte_path, bitOffset)

		//为剩余的path创建子结点
		nd_path = newRaxNode()
		nd_path.left = radixNode.left
		nd_path.right = radixNode.right
		nd_path.val = radixNode.val
		nd_path.bitLen = bitLastPath
		nd_path.bitVal = make([]byte, len(radixNode.bitVal[charIndex:]))
		copy(nd_path.bitVal, radixNode.bitVal[charIndex:])

		//将byte_path字节中的非公共部分分离出来,保存到子结点中
		if bitOffset != 0 {
			byte_tmp := CLEAR_BITS_LOW(byte_path, bitOffset)
			nd_path.bitVal[0] = byte_tmp
		}

		//修改当前结点，作为nd_path结点、nd_data结点的父结点
		//多申请一个子节，用于存储可能出现的不完整字节
		bit_val_old := radixNode.bitVal
		radixNode.left = nil
		radixNode.right = nil
		radixNode.val = nil
		radixNode.bitLen = radixNode.bitLen - bitLastPath // = bpos - (key_pos % 8)
		radixNode.bitVal = make([]byte, len(bit_val_old[0:charIndex])+1)
		copy(radixNode.bitVal, bit_val_old[0:charIndex])
		radixNode.bitVal = radixNode.bitVal[0 : len(radixNode.bitVal)-1]

		//将byte_path字节中的公共部分分离出来,保存到父结点
		if bitOffset != 0 {
			byte_tmp := CLEAR_BITS_HIGH(byte_path, 8-bitOffset)
			radixNode.bitVal = append(radixNode.bitVal, byte_tmp)
		}
	}

	//若path包含key，则将val赋值给radixNode结点
	if bitLastData == 0 {
		radixNode.val = val
	}
	if nd_data != nil {
		if bval_data == 0 {
			radixNode.left = nd_data
		} else {
			radixNode.right = nd_data
		}
	}
	if nd_path != nil {
		if bval_path == 0 {
			radixNode.left = nd_path
		} else {
			radixNode.right = nd_path
		}
	}
	return len(key) * 8, nil
}

func (radixNode *raxNode) pathCompare(data []byte, bbeg int) (bool, int) {
	bend := bbeg + int(radixNode.bitLen)
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
		byte_node := radixNode.bitVal[nci]
		if byte_data != byte_node {
			return false, len(data) * 8
		}

		bb += step
	}

	return true, bend
}

//将当前结点的子结点进行合并
//若当前结点只有一个子结点，并且当前结点是空结点，才可以进行合并操作
func (radixNode *raxNode) pathMerge(bpos int) bool {
	//若当前结点存在值，则不能合并
	if radixNode.val != nil {
		return false
	}

	//若当前结点有2个子结点，则不能合并
	if radixNode.left != nil && radixNode.right != nil {
		return false
	}

	//若当前结点没有子结点，则不能合并
	if radixNode.left != nil && radixNode.right != nil {
		return false
	}

	//获取当前结点的子结点
	child := radixNode.left
	if radixNode.right != nil {
		child = radixNode.right
	}

	//判断当前结点最后一个字节是否是完整的字节
	//若不是完整字节，需要与子结点的第一个字节进行合并
	if bpos%8 != 0 {
		char_len := len(radixNode.bitVal)
		char_last := radixNode.bitVal[char_len-1]
		char_0000 := child.bitVal[0]
		child.bitVal = child.bitVal[1:]
		radixNode.bitVal[char_len-1] = char_last | char_0000
	}

	//合并当前结点以及子结点
	radixNode.val = child.val
	radixNode.bitVal = append(radixNode.bitVal, child.bitVal...)
	radixNode.bitLen += child.bitLen
	radixNode.left = child.left
	radixNode.right = child.right

	return true
}

////打印结点信息，用于调试
func (radixNode *raxNode) getNodeInfo(bbeg int) string {
	buff := new(bytes.Buffer)

	bend := bbeg + int(radixNode.bitLen)
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
		byte_node := radixNode.bitVal[nci]

		//获取数据的当前字节以及循环步长
		step := 8
		if nci == 0 && obeg > 0 {
			step = 8 - obeg
		}
		if dci == cend && oend > 0 {
			step = oend
		}
		if cbeg == cend {
			step = int(radixNode.bitLen)
		}

		if step != 8 {
			buff.WriteString(fmt.Sprintf("(%08b:%d)", byte_node, byte_node))
		} else {
			buff.WriteByte(byte_node)
		}
		bb += step
	}

	if radixNode.val != nil {
		buff.WriteString(fmt.Sprintf("=%v", radixNode.val))
	}

	return buff.String()
}
