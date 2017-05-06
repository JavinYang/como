package comm

import (
	"sync"
	"time"
)

// 组织规划
type OrgPlanning struct {
	sync.Mutex
	pactRegisterName string
	remainingTime    time.Duration
	runningUpdates   map[chan interface{}]interface{}
	Leader           Leader
	MailBox          MailBox
}

// 初始化组织
func (this *OrgPlanning) init() {}

// 组织开始工作
func (this *OrgPlanning) start() {}

// 例行公事
func (this *OrgPlanning) routine() {}

// 带外消息
func (this *OrgPlanning) info() {}

// 组织解散
func (this *OrgPlanning) terminate() {}

// 组织领导
type Leader struct{}

// 添加事物循环处理
func (this *Leader) AddUpdate() {}

// 拒绝本次服务
func (this *Leader) DenialService() {}

// 获取超时时间
func (this *Leader) GetRemainingTime() {}

// 设置超时时间
func (this *Leader) SetRemainingTime() {}

// 解散组织
func (this *Leader) Dissolution() {}

// 邮箱
type MailBox struct {
	Address    chan *Mail  // 邮箱地址
	AddressMap *addressMap // 通讯录
}

// 写邮件
func (this *MailBox) Write() Draft {
	return Draft{}
}

// 读邮件
func (this *MailBox) Read() Mail {
	return Mail{}
}

// 通讯录
type addressMap struct {
	sync.Mutex
	addressMap map[chan *MailBox]*addressMap
}

// 添加好友
func (this *addressMap) AddFriend() bool {
	return true
}

// 删除好友
func (this *addressMap) DeleteFriend(mailBoxAddress chan *MailBox) {}

// 邮件
type Mail struct {
	SenderMailBoxAddress chan *Mail
	SenderName           string
	SendeeMailBoxAddress chan *Mail
	SendeeName           string
	Content              interface{}
	Remarks              map[string]interface{}
}

// 回复
func (this Mail) Reply() Draft {
	return Draft(this)
}

// 转发
func (this Mail) Forward() Draft {
	return Draft(this)
}

// 草稿
type Draft Mail

// 发送
func (this Draft) Send() bool {
	return true
}
