package dcache

import (
	"DCache/consistent"
	"sync"
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

	//与哈希环上所有的peer的http连接
	peers map[string]Peer
}

//创建一个HttpPool实例
func New(addr string) *HttpPool {
	return &HttpPool{
		selfAddr: addr,
	}
}

//设置peer
func (h *HttpPool) SetPeers(addrs ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hash = consistent.New(defaultReplicas, nil)
	for _, addr := range addrs {
		h.peers[addr] = &httpPeer{remoteBaseUrl: addr + defaultRoute}
	}
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
