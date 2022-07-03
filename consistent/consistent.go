package consistent

import (
	"hash/crc32"
	"sort"
	"strconv"
)

//hash函数，采用依赖注入的方式，允许替换成其他hash函数
type HashFunc func([]byte) uint32

//一致性哈希
type ConsistentHash struct {
	hashFunc HashFunc       //哈希函数
	replicas int            //虚拟节点个数
	ring     []int          //哈希环(存放的是虚拟节点)
	hashMap  map[int]string //虚拟节点与真实节点的映射值
}

//生成一致性哈希实例
func New(replicas int, hash HashFunc) *ConsistentHash {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	res := &ConsistentHash{
		hashFunc: hash,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	return res
}

//添加节点
func (c *ConsistentHash) AddNodes(realNodes ...string) {
	for _, node := range realNodes {
		//生成虚拟节点并添加到hash环上
		for i := 0; i < c.replicas; i++ {
			hashIndex := int(c.hashFunc([]byte(strconv.Itoa(i) + node)))
			c.ring = append(c.ring, hashIndex)
			c.hashMap[hashIndex] = node
		}
	}
	sort.Ints(c.ring)
}

//判断key落在哈希环上的哪个节点
func (c *ConsistentHash) GetNode(key string) string {
	hash := int(c.hashFunc([]byte(key)))

	//二分搜索找到第一个大于等于hash值的元素的index
	index := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hash
	})

	//index == len(c.ring)时交由第一个节点处理
	return c.hashMap[c.ring[index%len(c.ring)]]
}
