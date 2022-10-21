package lru

import (
	"reflect"
	"testing"
)

func TestLRUCache_Get(t *testing.T) {
	type args struct {
		key string
	}

	keyedLRU := NewLRUCache(3)
	keyedLRU.Put("test", "test")

	tests := []struct {
		name  string
		lru   *LRUCache
		args  args
		want  interface{}
		want1 bool
	}{
		{
			name: "NotFound",
			lru:  NewLRUCache(3),
			args: args{
				key: "test",
			},
			want:  nil,
			want1: false,
		},
		{
			name: "ConflictingKey",
			lru:  NewLRUCache(3),
			args: args{
				key: lru_tail,
			},
			want:  nil,
			want1: false,
		},
		{
			name: "Normal",
			lru:  keyedLRU,
			args: args{
				key: "test",
			},
			want:  "test",
			want1: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := tt.lru
			got, got1 := l.Get(tt.args.key)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Get() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("Get() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
