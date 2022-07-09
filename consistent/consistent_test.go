package consistent

import (
	"strconv"
	"testing"
)

func TestConsistentHash(t *testing.T) {
	consistentHash := New(3, func(key []byte) uint32 {
		res, _ := strconv.Atoi(string(key))
		return uint32(res)
	})

	//三个真实节点，分别命名为"2" "4"  "6"。那么根据hash函数，真实节点
	//与虚拟节点的映射为：
	//"2" : "2" "12" "22"
	//"4" : "4" "14" "24"
	//"6" : "6" "16" "26"
	consistentHash.AddNodes("2", "4", "6")
	cases := map[string]string{
		"2":  "2",
		"11": "2",
		"13": "4",
		"14": "4",
		"5":  "6",
		"15": "6",
		"17": "2",
		"25": "6",
		"28": "2",
	}
	for k, v := range cases {
		if res := consistentHash.GetNode(k); res != v {
			t.Errorf("for %s, want %s but get %s", k, v, res)
		}
	}
}
