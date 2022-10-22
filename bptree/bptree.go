package bptree

import "sync"

type BPTree struct {
	mu    sync.RWMutex
	ktype int
	root  *bpNode
	width int
	halfw int
}

func NewBPTree(width int) *BPTree {
	if width < 3 {
		width = 3
	}

	tree := &BPTree{}
	tree.root = newLeafNode(width)
	tree.width = width
	tree.halfw = (tree.width + 1) / 2
	return tree
}

func (t *BPTree) Get(key int64) interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node := t.root
	for i := 0; i < len(node.nodes); i++ {
		if key <= node.nodes[i].maxKey {
			node = node.nodes[i]
			i = 0
		}
	}

	if len(node.nodes) > 0 {
		return nil
	}

	for i := 0; i < len(node.items); i++ {
		if node.items[i].key == key {
			return node.items[i].value
		}
	}
	return nil
}

func (t *BPTree) Set(key int64, value interface{}) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.setValue(nil, t.root, key, value)
}

func (t *BPTree) GetData() map[int64]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.getData(t.root)
}

func (t *BPTree) getData(node *bpNode) map[int64]interface{} {
	data := make(map[int64]interface{})

	for {
		if len(node.nodes) > 0 {
			for i := 0; i < len(node.nodes); i++ {
				data[node.nodes[i].maxKey] = t.getData(node.nodes[i])
			}
			break
		} else {
			for i := 0; i < len(node.items); i++ {
				data[node.items[i].key] = node.items[i].value
			}
			break
		}
	}
	return data
}

func (t *BPTree) splitNode(node *bpNode) *bpNode {
	if len(node.nodes) > t.width {
		halfw := t.width/2 + 1
		node2 := newIndexNode(t.width)
		node2.nodes = append(node2.nodes, node.nodes[halfw:len(node.nodes)]...)
		node2.maxKey = node2.nodes[len(node2.nodes)-1].maxKey

		node.nodes = node.nodes[0:halfw]
		node.maxKey = node.nodes[len(node.nodes)-1].maxKey
		return node2
	} else if len(node.items) > t.width {
		halfw := t.width/2 + 1
		node2 := newLeafNode(t.width)
		node2.items = append(node2.items, node.items[halfw:len(node.items)]...)
		node2.maxKey = node2.items[len(node2.items)-1].key

		node.next = node2
		node.items = node.items[0:halfw]
		node.maxKey = node.items[len(node.items)-1].key
		return node2
	}
	return nil
}

func (t *BPTree) setValue(parent *bpNode, node *bpNode, key int64, value interface{}) {
	for i := 0; i < len(node.nodes); i++ {
		if key <= node.nodes[i].maxKey || i == len(node.nodes)-1 {
			t.setValue(node, node.nodes[i], key, value)
			break
		}
	}

	if len(node.nodes) < 1 {
		node.setValue(key, value)
	}

	newNode := t.splitNode(node)
	if newNode != nil {
		if parent == nil {
			parent = newIndexNode(t.width)
			parent.addChild(node)
			t.root = parent
		}
		parent.addChild(newNode)
	}
}

func (t *BPTree) itemMoveOrMerge(parent *bpNode, node *bpNode) {
	var node1 *bpNode = nil
	var node2 *bpNode = nil
	for i := 0; i < len(parent.nodes); i++ {
		if parent.nodes[i] == node {
			if i < len(parent.nodes)-1 {
				node2 = parent.nodes[i+1]
			} else if i > 0 {
				node1 = parent.nodes[i-1]
			}
			break
		}
	}

	//将左侧结点的记录移动到删除结点
	if node1 != nil && len(node1.items) > t.halfw {
		item := node1.items[len(node1.items)-1]
		node1.items = node1.items[0 : len(node1.items)-1]
		node1.maxKey = node1.items[len(node1.items)-1].key
		node.items = append([]bpItem{item}, node.items...)
		return
	}

	//将右侧结点的记录移动到删除结点
	if node2 != nil && len(node2.items) > t.halfw {
		item := node2.items[0]
		node2.items = node1.items[1:]
		node.items = append(node.items, item)
		node.maxKey = node.items[len(node.items)-1].key
		return
	}

	//与左侧结点进行合并
	if node1 != nil && len(node1.items)+len(node.items) <= t.width {
		node1.items = append(node1.items, node.items...)
		node1.next = node.next
		node1.maxKey = node1.items[len(node1.items)-1].key
		parent.deleteChild(node)
		return
	}

	//与右侧结点进行合并
	if node2 != nil && len(node2.items)+len(node.items) <= t.width {
		node.items = append(node.items, node2.items...)
		node.next = node2.next
		node.maxKey = node.items[len(node.items)-1].key
		parent.deleteChild(node2)
		return
	}
}

func (t *BPTree) childMoveOrMerge(parent *bpNode, node *bpNode) {
	if parent == nil {
		return
	}

	//获取兄弟结点
	var node1 *bpNode = nil
	var node2 *bpNode = nil
	for i := 0; i < len(parent.nodes); i++ {
		if parent.nodes[i] == node {
			if i < len(parent.nodes)-1 {
				node2 = parent.nodes[i+1]
			} else if i > 0 {
				node1 = parent.nodes[i-1]
			}
			break
		}
	}

	//将左侧结点的子结点移动到删除结点
	if node1 != nil && len(node1.nodes) > t.halfw {
		item := node1.nodes[len(node1.nodes)-1]
		node1.nodes = node1.nodes[0 : len(node1.nodes)-1]
		node.nodes = append([]*bpNode{item}, node.nodes...)
		return
	}

	//将右侧结点的子结点移动到删除结点
	if node2 != nil && len(node2.nodes) > t.halfw {
		item := node2.nodes[0]
		node2.nodes = node1.nodes[1:]
		node.nodes = append(node.nodes, item)
		return
	}

	if node1 != nil && len(node1.nodes)+len(node.nodes) <= t.width {
		node1.nodes = append(node1.nodes, node.nodes...)
		parent.deleteChild(node)
		return
	}

	if node2 != nil && len(node2.nodes)+len(node.nodes) <= t.width {
		node.nodes = append(node.nodes, node2.nodes...)
		parent.deleteChild(node2)
		return
	}
}

func (t *BPTree) deleteItem(parent *bpNode, node *bpNode, key int64) {
	for i := 0; i < len(node.nodes); i++ {
		if key <= node.nodes[i].maxKey {
			t.deleteItem(node, node.nodes[i], key)
			break
		}
	}

	if len(node.nodes) < 1 {
		//删除记录后若结点的子项<m/2，则从兄弟结点移动记录，或者合并结点
		node.deleteItem(key)
		if len(node.items) < t.halfw {
			t.itemMoveOrMerge(parent, node)
		}
	} else {
		//若结点的子项<m/2，则从兄弟结点移动记录，或者合并结点
		node.maxKey = node.nodes[len(node.nodes)-1].maxKey
		if len(node.nodes) < t.halfw {
			t.childMoveOrMerge(parent, node)
		}
	}
}

func (t *BPTree) Remove(key int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.deleteItem(nil, t.root, key)
}
