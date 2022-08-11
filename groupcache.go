package cache

import (
	"fmt"
	"sync"
	"time"

	"github.com/hollowdjj/course-selecting-sys/cache/pb"
	"github.com/hollowdjj/course-selecting-sys/cache/singleshot"
	"github.com/hollowdjj/course-selecting-sys/pkg/logger"
	"github.com/sirupsen/logrus"

	"github.com/bits-and-blooms/bloom/v3"
)

//For each group cache one must implement Getter interface in case to load cache
//when no cache is found either in local cache pool and peer cache pools.
type Getter interface {
	Get(key string) ([]byte, error)
}

//A function type, so that Getter can be a function
type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

//cache query option
type Option struct {
	FromLocal  bool
	FromPeer   bool
	FromGetter bool
	TTL        time.Duration //unit: second
}

var (
	rw      sync.RWMutex
	groups  = make(map[string]*GroupCache)
	runMode string

	DefaultOption = Option{true, true, true, 300}
)

//GroupCache stores cache that can be put in the same gruop, eg: student,course
type GroupCache struct {
	//group name
	name string

	//max bytes a GroupCache can hold
	maxBytes int64

	//getter, passed from user
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

//get cache from GroupCache according to the key. Value might be empty according
//to cache query option
func (g *GroupCache) Get(key string, opt Option) (Value, error) {
	if key == "" {
		msg := "key requied inorder to get cache"
		logger.GetInstance().Errorln("msg")
		return Value{}, fmt.Errorf(msg)
	}

	logger.GetInstance().WithFields(logrus.Fields{
		"group": g.name,
		"key":   key,
		"opt":   opt,
	}).Infoln("set cache query option succ")

	//bloom filter
	if g.enableBloomFilter && !g.bloom.Test([]byte(key)) {
		logger.GetInstance().WithFields(logrus.Fields{
			"group": g.name,
			"key":   key,
		}).Infoln("key filtered by bloom filter")
		return Value{}, fmt.Errorf("key [%v] filtered by bloom filter", key)
	}

	//look up in local cache first
	if opt.FromLocal {
		if val, hit := g.lookupLocalCache(key); hit {
			logger.GetInstance().WithFields(logrus.Fields{
				"group": g.name,
				"key":   key,
			}).Infoln("get cache from local cache succ")
			return val, nil
		}
	}

	//if not find in local cache, load cache
	if !opt.FromPeer && !opt.FromGetter {
		return Value{}, nil
	}
	val, err := g.loadCache(key, opt)
	if err != nil {
		return Value{}, err
	}
	return val, nil
}

//look up in local cache
func (g *GroupCache) lookupLocalCache(key string) (Value, bool) {
	if val, hit := g.mainCache.get(key); hit {
		return val, true
	}
	if val, hit := g.hotCache.get(key); hit {
		return val, true
	}
	return Value{}, false
}

//get cache from a peer or Getter
func (g *GroupCache) loadCache(key string, opt Option) (Value, error) {
	fromPeer := false
	val, err := g.shot.Do(key, func() (val interface{}, err error) {
		//get from peer
		if opt.FromPeer && g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				fromPeer = true
				return g.getFromPeer(peer, key)
			}
		}
		//get from Getter
		if opt.FromGetter {
			return g.getFromGetter(key)
		}
		return
	})

	if err != nil {
		logger.GetInstance().WithField("err", err).Errorln("fail to load cache")
		return Value{}, err
	}

	//write cache to hotCache if from peer, else write to mainCache
	res := val.(Value)
	if len(res.ByteSlice()) != 0 {
		if fromPeer {
			g.hotCache.add(key, res, opt.TTL)
		} else {
			g.mainCache.add(key, res, opt.TTL)
		}
	}

	//check overflow
	for g.mainCache.nbytes+g.hotCache.nbytes > g.maxBytes {
		g.hotCache.removeLeastUsed()
	}

	return res, nil
}

//get cache from peer
func (g *GroupCache) getFromPeer(peer Peer, key string) (Value, error) {
	req := &pb.GetRequest{Group: g.name, Key: key}
	resp := &pb.GetResponse{}
	err := peer.Get(req, resp)
	if err != nil {
		logger.GetInstance().WithFields(logrus.Fields{
			"group": g.name,
			"key":   key,
			"peer":  peer.Addr(),
			"err":   err,
		}).Errorln("get cache from peer failed")
		return Value{}, fmt.Errorf("get cache from peer [%v] failed: %v", peer.Addr(), err)
	}
	logger.GetInstance().WithFields(logrus.Fields{
		"group": g.name,
		"key":   key,
		"peer":  peer.Addr(),
	}).Infoln("get cache from peer succ")

	return Value{b: resp.GetValue()}, nil
}

//get from Getter
func (g *GroupCache) getFromGetter(key string) (Value, error) {
	if g.getter == nil {
		return Value{}, nil
	}
	bytes, err := g.getter.Get(key)
	if err != nil {
		logger.GetInstance().WithFields(logrus.Fields{
			"group": g.name,
			"key":   key,
			"err":   err,
		}).Infoln("get cache from getter failed")
		return Value{}, err
	}

	logger.GetInstance().WithFields(logrus.Fields{
		"group": g.name,
		"key":   key,
	}).Infoln("get cache from getter succ")
	return Value{b: bytes}, nil
}

//Add cache, if key already exist, its value will be update to data
func (g *GroupCache) Add(key string, data []byte, ttl time.Duration) {
	g.mainCache.add(key, Value{data}, ttl)
}

//Del cache,if key is not exist nothing will happen
func (g *GroupCache) Del(key string) {
	g.mainCache.del(key)
	g.hotCache.del(key)
}

//get a new group cache instance, concurrency safe
func NewGroupCache(name string, maxBytes int64, getter Getter) *GroupCache {
	res := &GroupCache{
		name:     name,
		maxBytes: maxBytes,
		getter:   getter,
		shot:     &singleshot.Shots{},
	}
	rw.Lock()
	defer rw.Unlock()
	if ret, hit := groups[name]; hit {
		return ret
	}
	groups[name] = res
	return res
}

//get instance of group cache according to group name, concurrency safe
func GetGroupCache(name string) *GroupCache {
	rw.RLock()
	defer rw.RUnlock()
	g := groups[name]
	return g
}
