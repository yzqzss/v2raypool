package v2raypool

import (
	"fmt"
	// "os/exec"
	"sort"
	"time"

	"github.com/v2fly/v2ray-core/v5/common/net"
)

type ProxyNode struct {
	Index, LocalPort int
	// cmd               *exec.Cmd
	Id, localAddr   string
	RemoteAddr      string `json:"remote_addr"`
	Title, Protocol string
	TestUrl         string
	Speed           time.Duration
	TestAt          time.Time
	v2rayNode       V2rayNode
	status          int
}

func NewProxyNodeByV2ray(vnd V2rayNode) *ProxyNode {
	n := &ProxyNode{}
	n.SetV2ray(vnd)
	return n
}

func (p *ProxyNode) GetId() string {
	if p.Id != "" {
		return p.Id
	}
	p.Id = p.RemoteAddr + ":" + p.v2rayNode.Id
	return p.Id
}
func (p *ProxyNode) SetV2ray(n V2rayNode) *ProxyNode {
	p.RemoteAddr = fmt.Sprintf("%s:%v", n.Add, n.Port)
	p.Id = p.RemoteAddr + ":" + n.Id
	p.Title = n.Ps
	p.Protocol = n.Protocol
	p.v2rayNode = n
	return p
}

func (p *ProxyNode) AddToPool(c *V2rayApiClient) error {
	tag := getProxyNodeTag(p.Index)
	cf := getConf()
	protcl := cf.GetHttpProxyProtocol()
	if protcl == "socks5" {
		protcl = "socks"
	}
	err := c.AddInbound(net.Port(p.LocalPort), tag, protcl)
	if err != nil {
		return err
	}
	err = c.AddOutboundByV2rayNode(p.v2rayNode, tag)
	if err != nil {
		return err
	}
	p.status = 1
	return err
}

func (p *ProxyNode) Remove(c *V2rayApiClient, tag string) error {
	if tag == "" {
		tag = p.GetId()
	}
	err := c.RemoveOutbound(tag)
	if err != nil {
		return err
	}
	err = c.RemoveInbound(tag)
	if err != nil {
		return err
	}
	p.status = 0
	return err
}

func (p ProxyNode) IsRunning() bool {
	return p.status == 1
}

// IsOk 查看测速是否超过有效期。默认24小时
func (p ProxyNode) IsOk() bool {
	return time.Since(p.TestAt) < time.Hour*24
}

// // {"add":"jp6.xxx.top","host":"","id":"0999AE93-1330-4A75-DBC1-0DD545F7DD60","net":"ws","path":"","port":"41444","ps":"xxx-v2-JP-Tokyo6(1)","tls":"","v":2,"aid":0,"type":"none"}
// protocol, add, port id, net

type ProxyNodes []ProxyNode

func (s ProxyNodes) Len() int           { return len(s) }
func (s ProxyNodes) Less(i, j int) bool { return s[i].Speed < s[j].Speed }
func (s ProxyNodes) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s *ProxyNodes) SortBySpeed() {
	sort.Sort(s)
}
