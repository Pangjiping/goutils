package bptree

import (
	"encoding/json"
	"github.com/Pangjiping/goutils/utils"
	"math/rand"
	"testing"
)

func TestBPTree(t *testing.T) {
	tree := NewBPTree(4)

	tree.Set(10, 1)
	tree.Set(23, 1)
	tree.Set(33, 1)
	tree.Set(35, 1)
	tree.Set(15, 1)
	tree.Set(16, 1)
	tree.Set(17, 1)
	tree.Set(19, 1)
	tree.Set(20, 1)

	t.Logf("b+ tree data:%v", tree.GetData())

	tree.Remove(23)

	t.Log(tree.Get(10))
	t.Log(tree.Get(15))
	t.Log(tree.Get(20))

	data, _ := json.MarshalIndent(tree.GetData(), "", "    ")
	t.Log(utils.Bytes2String(data))
}

func TestBPTRand(t *testing.T) {
	bpt := NewBPTree(3)

	for i := 0; i < 12; i++ {
		key := rand.Int()%20 + 1
		t.Log(key)
		bpt.Set(int64(key), key)
	}

	data, _ := json.MarshalIndent(bpt.GetData(), "", "    ")
	t.Log(string(data))
}
