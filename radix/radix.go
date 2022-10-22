package radix

import (
	"encoding/json"
	"github.com/Pangjiping/goutils/utils"
)

type Radix struct {
	root [256]*raxNode
}

//添加元素
func (this *Radix) Set(key string, val interface{}) {
	data := []byte(key)
	root := this.root[data[0]]
	if root == nil {
		//没有根节点，则创建一个根节点
		root = newRaxNode()
		root.val = val
		root.bitVal = make([]byte, len(data))
		copy(root.bitVal, data)
		root.bitLen = len(data) * 8
		this.root[data[0]] = root
		return
	} else if root.val == nil && root.left == nil && root.right == nil {
		//只有一个根节点，并且是一个空的根节点，则直接赋值
		root.val = val
		root.bitVal = make([]byte, len(data))
		copy(root.bitVal, data)
		root.bitLen = len(data) * 8
		this.root[data[0]] = root
		return
	}

	cur := root
	blen := len(data) * 8
	for bpos := 0; bpos < blen && cur != nil; {
		bpos, cur = cur.pathSplit(data, bpos, val)
	}
}

//查找元素
func (this *Radix) Get(key string) interface{} {
	data := []byte(key)
	blen := len(data) * 8
	cur := this.root[data[0]]

	var iseq bool
	for bpos := 0; bpos < blen && cur != nil; {
		iseq, bpos = cur.pathCompare(data, bpos)
		if iseq == false {
			return nil
		}
		if bpos >= blen {
			return cur.val
		}

		byte_data := data[bpos/8]
		bit_pos := GET_BIT(byte_data, bpos%8)
		if bit_pos == 0 {
			cur = cur.left
		} else {
			cur = cur.right
		}
	}

	return nil
}

//删除元素
//只能删除叶子结点，不能删除根节点
//删除叶子结点后，若parent结点只有一个子结点，则将parent结点与子结点合并
func (this *Radix) Delete(key string) {
	data := utils.String2Bytes(key)
	blen := len(data) * 8
	cur := this.root[data[0]]

	var iseq bool
	var parent *raxNode = nil
	for bpos := 0; bpos < blen && cur != nil; {
		iseq, bpos = cur.pathCompare(data, bpos)
		if iseq == false {
			return
		}

		if bpos >= blen {
			//将当前结点修改为空结点
			//若parent是根节点，不能删除
			cur.val = nil
			if parent == nil {
				return
			}

			//当前结点是叶子结点，先将当前结点删除，并将当前结点指向父结点
			if cur.left == nil && cur.right == nil {
				if parent.left == cur {
					parent.left = nil
				} else if parent.right == cur {
					parent.right = nil
				}
				bpos -= int(cur.bitLen)
				cur = parent
			}

			//尝试将当前结点与当前结点的子节点进行合并
			cur.pathMerge(bpos)
			return
		}

		byte_data := data[bpos/8]
		bit_pos := GET_BIT(byte_data, bpos%8)
		if bit_pos == 0 {
			parent = cur
			cur = cur.left
		} else {
			parent = cur
			cur = cur.right
		}
	}
}

//递归获取数据，用于调试
func (this *Radix) getItems(cur *raxNode, bpos int, key []byte, items []stringKV) []stringKV {
	//备份key数据
	key_len := len(key)
	var key_last byte
	if key_len > 0 {
		key_last = key[key_len-1]
	}

	//合并key数据
	if bpos%8 != 0 {
		key = key[0 : key_len-1]
		key = append(key, key_last|cur.bitVal[0])
		key = append(key, cur.bitVal[1:]...)
	} else {
		key = append(key, cur.bitVal...)
	}
	bpos += int(cur.bitLen)

	//将value以及可以加入结果集
	if cur.val != nil {
		item := stringKV{string(key), cur.val}
		items = append(items, item)
	}

	if cur.left != nil {
		items = this.getItems(cur.left, bpos, key, items)
	}
	if cur.right != nil {
		items = this.getItems(cur.right, bpos, key, items)
	}

	//恢复key数据
	key = key[0:key_len]
	if key_len > 0 {
		key[key_len-1] = key_last
	}
	return items
}

//获取数据，用于调试
func (this *Radix) GetItems() []stringKV {
	items := make([]stringKV, 0)
	key := make([]byte, 0)
	for i := 0; i < 255; i++ {
		cur := this.root[i]
		if cur == nil {
			continue
		}
		items = this.getItems(cur, 0, key, items)
	}
	return items
}

//递归打印结点信息，用于调试
func (this *Radix) getNodesInfo(cur *raxNode, pos int, data map[string]interface{}) {
	data["info"] = cur.getNodeInfo(pos)
	pos += int(cur.bitLen)

	if cur.left != nil {
		tmp := make(map[string]interface{})
		data["left"] = tmp
		this.getNodesInfo(cur.left, pos, tmp)
	}

	if cur.right != nil {
		tmp := make(map[string]interface{})
		data["right"] = tmp
		this.getNodesInfo(cur.right, pos, tmp)
	}
}

//打印结点信息，用于调试
func (this *Radix) GetNodesInfo(cc byte) string {
	cur := this.root[cc]
	data_root := make(map[string]interface{})
	this.getNodesInfo(cur, 0, data_root)
	ret, _ := json.MarshalIndent(data_root, "", "    ")
	return string(ret)
}
