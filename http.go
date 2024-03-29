package cache

import (
	"sync"

	"github.com/hollowdjj/course-selecting-sys/cache/consistent"
)

const (
	defaultReplicas = 50
	defaultRoute    = "/_dcache"
)

//Http连接池，保存有与哈希环上所有其他节点的http连接
type HttpPool struct {
	mu sync.Mutex

	//本机地址 eg:xx.xx.xxx.xx:8000
	selfAddr string

	//一致性哈希
	hash *consistent.ConsistentHash

	//与所有真实节点的连接
	peers map[string]Peer
}

//创建一个HttpPool实例。selfAddr eg:127.0.0.1:8000
func NewHttpPool(selfAddr string) *HttpPool {
	return &HttpPool{
		selfAddr: selfAddr,
	}
}

//init consistent hash if it is not initialized and add peers.
//return all current realworld nodes on hash ring. Concurrency safe
func (h *HttpPool) AddPeers(addrs ...string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	//lazy initialization
	if h.hash == nil {
		h.hash = consistent.New(defaultReplicas, nil)
	}
	if h.peers == nil {
		h.peers = make(map[string]Peer)
	}

	for _, addr := range addrs {
		h.peers[addr] = &httpPeer{remoteBaseUrl: "http://" + addr + defaultRoute}
	}
	h.hash.AddNodes(addrs...)
	var res []string
	for k := range h.peers {
		res = append(res, k)
	}
	return res
}

//return all peers, concurrency safe
func (h *HttpPool) GetPeers() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	var res []string
	for k := range h.peers {
		res = append(res, k)
	}
	return res
}

//delete peer
func (h *HttpPool) DelPeer(host string) []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.peers, host)
	h.hash.DelNode(host)
	var res []string
	for k := range h.peers {
		res = append(res, k)
	}
	return res
}

//设置一致性哈希
func (h *HttpPool) SetConsistentHash(hash *consistent.ConsistentHash) {
	h.hash = hash
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
