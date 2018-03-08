package como

import (
	"net"
)

type nodeCenter struct {
	myAddr net.TCPAddr                          // 我的TCP地址
	nodes  map[string]map[*nodeAddress]struct{} // 组名 节点邮箱地址
}

type nodeAddress struct {
	mailBoxAddr MailBoxAddress
	ipPort      string // ip端口
}

func (this *nodeCenter) Init(pars ...interface{}) {

}

func (this *nodeCenter) start() {
	// 循环连接所有可连接组
	// 循环接收新接入请求
}

func (this *nodeCenter) close() {

}
