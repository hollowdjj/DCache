package DCache

import (
	"fmt"
	"sync"

	"github.com/hollowdjj/DCache/pb"
	"github.com/hollowdjj/DCache/singleshot"
	"github.com/sirupsen/logrus"

	"github.com/bits-and-blooms/bloom/v3"
)

//Getter接口。由用户实现，进而在缓存不存在时获取缓存数据并添加至缓存
type Getter interface {
	Get(key string) ([]byte, error)
}

//定义一个函数类型，从而函数也可以传参给Getter接口
type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

var (
	rw     sync.RWMutex
	groups = make(map[string]*GroupCache)
)

type GroupCache struct {
	//group名
	name string

	//mainCache+hotCache的内存上限
	maxBytes int64

	//每一个groupcache都必须有getter，从而在缓存不存在时获取缓存
	getter Getter

	//通过一致性哈希计算后，落在本节点的缓存
	mainCache cache

	//一致性哈希后不应由本节点保存的缓存，但本节点又经常收到相关请求。
	//为了避免网络开销，保存一个副本。
	hotCache cache

	//分布式节点集
	peers PeerPicker

	//避免缓存击穿
	shot *singleshot.Shots

	//是否使用布隆过滤器
	enableBloomFilter bool

	//布隆过滤器，防止缓存穿透
	bloom *bloom.BloomFilter
}

//注册peerpicker
func (g *GroupCache) RegisterPeerPicker(picker PeerPicker) {
	g.peers = picker
}

//创建并激活一个存放大约n个元素，误判率为fp的布隆过滤器
func (g *GroupCache) EnableBloomFilter(n uint, fp float64) {
	g.bloom = bloom.NewWithEstimates(n, fp)
	g.enableBloomFilter = true
}

//从GroupCache中获取缓存值
func (g *GroupCache) GetCacheValue(key string) (Value, error) {
	if key == "" {
		return Value{}, fmt.Errorf("key requied inorder to get cache")
	}

	if g.enableBloomFilter && !g.bloom.Test([]byte(key)) {
		return Value{}, fmt.Errorf("key [%v] is not in bloom filter", key)
	}

	//先查本地缓存
	if val, hit := g.LookupLocalCache(key); hit {
		if enableLogger {
			GetLogger().WithFields(logrus.Fields{
				"group": g.name,
				"key":   key,
			}).Info("get cache from local cache")
		}
		return val, nil
	}

	//本地缓冲中没有，那么加载缓存至本地缓存
	val, err := g.loadCache(key)
	if err != nil {
		return Value{}, err
	}
	return val, nil
}

//查询本地缓存
func (g *GroupCache) LookupLocalCache(key string) (Value, bool) {
	if val, hit := g.mainCache.get(key); hit {
		return val, true
	}
	if val, hit := g.hotCache.get(key); hit {
		return val, true
	}
	return Value{}, false
}

//从远端或者通过Getter获取缓存，使用singleshot防止缓存击穿
func (g *GroupCache) loadCache(key string) (Value, error) {
	val, err := g.shot.Do(key, func() (interface{}, error) {
		//从peer获取缓存
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				return g.getFromPeer(peer, key)
			}
		}
		//从getter获取缓存
		return g.getFromGetter(key)
	})

	if err != nil {
		return Value{}, err
	}

	//将缓存写入mainCache以及hotCache
	res := val.(Value)
	g.mainCache.add(key, res)
	g.hotCache.add(key, res)
	return res, nil
}

//查询远端缓存
func (g *GroupCache) getFromPeer(peer Peer, key string) (Value, error) {
	req := &pb.GetRequest{Group: g.name, Key: key}
	resp := &pb.GetResponse{}
	err := peer.Get(req, resp)
	if err != nil {
		return Value{}, fmt.Errorf("Get cache from peer [%v] failed: %v", peer.Addr(), err)
	}
	if enableLogger {
		GetLogger().WithFields(logrus.Fields{
			"remote": peer.Addr(),
			"group":  g.name,
			"key":    key,
		}).Info("get cache from remote")
	}
	return Value{b: resp.GetValue()}, nil
}

//通过getter获取缓存
func (g *GroupCache) getFromGetter(key string) (Value, error) {
	if g.getter == nil {
		return Value{}, nil
	}
	bytes, err := g.getter.Get(key)
	if err != nil {
		return Value{}, err
	}
	if enableLogger {
		GetLogger().WithFields(logrus.Fields{
			"group": g.name,
			"key":   key,
		}).Info("get cache from getter")
	}
	return Value{b: bytes}, nil
}

//生成一个GroupCache实例
func NewGroupCache(name string, maxBytes int64, getter Getter) *GroupCache {
	res := &GroupCache{
		name:     name,
		maxBytes: maxBytes,
		getter:   getter,
		shot:     &singleshot.Shots{},
	}
	rw.Lock()
	defer rw.Unlock()
	groups[name] = res
	return res
}

//获取GroupCache实例
func GetGroupCache(name string) *GroupCache {
	rw.RLock()
	defer rw.RUnlock()
	g := groups[name]
	return g
}
