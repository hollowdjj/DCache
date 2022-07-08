package dcache

import (
	"fmt"
	"github/hollowdjj/DCache/pb"
	"io/ioutil"
	"net/http"
	"net/url"

	"google.golang.org/protobuf/proto"
)

type PeerPicker interface {
	PickPeer(key string) (Peer, bool)
}

//抽象的peer节点(可以是http客户端，也可以是一个rpc调用)
//只要实现了Peer接口就可以认为是一个peer节点
type Peer interface {
	Get(*pb.GetRequest, *pb.GetResponse) error
	Addr() string
}

//http实现的peer
type httpPeer struct {
	remoteBaseUrl string //eg: http://xx.xxx.xxx.xx:8000/_dcache
}

func (h *httpPeer) Get(req *pb.GetRequest, resp *pb.GetResponse) error {
	//拼接完整url
	url := fmt.Sprint("%v?group=%v&key=%v", h.remoteBaseUrl,
		url.QueryEscape(req.GetGroup()), url.QueryEscape(req.GetKey()))

	//发送http请求
	response, err := http.Get(url)
	defer response.Body.Close()
	if err != nil {
		return err
	}
	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	if err = proto.Unmarshal(bytes, resp); err != nil {
		return fmt.Errorf("Decode protobuf response failed: %v", err)
	}

	return nil
}

func (h *httpPeer) Addr() string {
	return h.Addr()
}
