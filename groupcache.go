package dcache

//Getter接口。由用户实现，进而在缓存不存在时获取缓存数据并添加至缓存
type Getter interface {
	Get(key string) ([]byte, error)
}

//定义一个函数类型，从而函数也可以传参给Getter接口
type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g.Get(key)
}

type GroupCache struct {
	name string

	getter Getter

	//通过一致性哈希计算后，落在本节点的缓存
	mainCache cache

	//一致性哈希后不应由本节点保存的缓存，但本节点又经常收到相关请求。
	//为了避免网络开销，保存一个副本。
	hotCache cache
}
