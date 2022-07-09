package DCache

import (
	"sync"

	"github.com/hollowdjj/DCache/consistent"
)

const (
	defaultReplicas = 50
	defaultRoute    = "/_dcache"
)

//Http连接池，保存有与哈希环上所有其他节点的http连接
type HttpPool struct {
	mu sync.Mutex

	//本机地址 eg:http://xx.xx.xxx.xx:8000
	selfAddr string

	//一致性哈希
	hash *consistent.ConsistentHash

	//与所有真实节点的连接
	peers map[string]Peer
}

//创建一个HttpPool实例
func NewHttpPool(selfAddr string) *HttpPool {
	return &HttpPool{
		selfAddr: selfAddr,
	}
}

//设置peer
func (h *HttpPool) SetPeers(addrs ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hash = consistent.New(defaultReplicas, nil)
	h.peers = make(map[string]Peer)
	for _, addr := range addrs {
		h.peers[addr] = &httpPeer{remoteBaseUrl: addr + defaultRoute}
	}
	h.hash.AddNodes(addrs...)
}

//根据key的哈希值选择节点。
//当key经过hash后，落在hash环上的节点存在且不是本机时，返回peer，true
//否则返回nil,false
func (h *HttpPool) PickPeer(key string) (Peer, bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if node := h.hash.GetNode(key); node != "" && node != h.selfAddr {
		return h.peers[node], true
	}

	return nil, false
}
