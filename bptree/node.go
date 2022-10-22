package bptree

type bpItem struct {
	key   int64
	value interface{}
}

type stringKV struct {
	key   string
	value interface{}
}

type bpNode struct {
	maxKey int64
	nodes  []*bpNode
	items  []bpItem
	next   *bpNode
}

func newLeafNode(width int) *bpNode {
	node := &bpNode{}
	node.items = make([]bpItem, width+1)
	node.items = node.items[0:0]
	return node
}

func newIndexNode(width int) *bpNode {
	node := &bpNode{}
	node.nodes = make([]*bpNode, width+1)
	node.nodes = node.nodes[0:0]
	return node
}

func (node *bpNode) findItem(key int64) int {
	num := len(node.items)
	for i := 0; i < num; i++ {
		if node.items[i].key > key {
			return -1
		} else if node.items[i].key == key {
			return i
		}
	}
	return -1
}

func (node *bpNode) setValue(key int64, value interface{}) {
	item := bpItem{
		key:   key,
		value: value,
	}

	num := len(node.items)
	if num < 1 {
		node.items = append(node.items, item)
		node.maxKey = item.key
		return
	} else if key < node.items[0].key {
		node.items = append([]bpItem{item}, node.items...)
		return
	} else if key > node.items[num-1].key {
		node.items = append(node.items, item)
		node.maxKey = item.key
		return
	}

	for i := 0; i < num; i++ {
		if node.items[i].key > key {
			node.items = append(node.items, bpItem{})
			copy(node.items[i+1:], node.items[i:])
			node.items[i] = item
			return
		} else if node.items[i].key == key {
			node.items[i] = item
			return
		}
	}
}

func (node *bpNode) addChild(child *bpNode) {
	num := len(node.nodes)
	if num < 1 {
		node.nodes = append(node.nodes, child)
		node.maxKey = child.maxKey
		return
	} else if child.maxKey < node.nodes[0].maxKey {
		node.nodes = append([]*bpNode{child}, node.nodes...)
		return
	} else if child.maxKey > node.nodes[num-1].maxKey {
		node.nodes = append(node.nodes, child)
		node.maxKey = child.maxKey
		return
	}

	for i := 0; i < num; i++ {
		if node.nodes[i].maxKey > child.maxKey {
			node.nodes = append(node.nodes, nil)
			copy(node.nodes[i+1:], node.nodes[i:])
			node.nodes[i] = child
			return
		}
	}
}

func (node *bpNode) deleteItem(key int64) bool {
	num := len(node.items)

	for i := 0; i < num; i++ {
		if node.items[i].key > key {
			return false
		} else if node.items[i].key == key {
			copy(node.items[i:], node.items[i+1:])
			node.items = node.items[0 : len(node.items)-1]
			node.maxKey = node.items[len(node.items)-1].key
			return true
		}
	}
	return false
}

func (node *bpNode) deleteChild(child *bpNode) bool {
	num := len(node.nodes)
	for i := 0; i < num; i++ {
		if node.nodes[i] == child {
			copy(node.nodes[i:], node.nodes[i+1:])
			node.nodes = node.nodes[0 : len(node.nodes)-1]
			node.maxKey = node.nodes[len(node.nodes)-1].maxKey
			return true
		}
	}
	return false
}
