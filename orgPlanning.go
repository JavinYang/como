package comm

import (
	"sync"
	"time"
)

// 组织规划
type OrgPlanning struct {
	lock             sync.Mutex
	pactRegisterName *string
	remainingTime    *time.Duration
	runningUpdates   *map[chan interface{}]interface{}
	T_T              *leader
	MailBox          *mailBox
}

// 初始化组织
func (this *OrgPlanning) init(pactRegisterName *string, remainingTime *time.Duration, runningUpdates *map[chan interface{}]interface{}, t_t *leader, mailBox *mailBox) {
}

// 组织开始工作
func (this *OrgPlanning) Start() {}

// 例行公事
func (this *OrgPlanning) Routine() {}

// 带外消息
func (this *OrgPlanning) Info() {}

// 组织解散
func (this *OrgPlanning) Terminate() {}

// 组织领导
type leader struct{}

// 添加事物循环处理
func (this *leader) AddUpdate() {}

// 拒绝本次服务
func (this *leader) DenialService() {}

// 获取超时时间
func (this *leader) GetRemainingTime() {}

// 设置超时时间
func (this *leader) SetRemainingTime() {}

// 邮箱
type mailBox struct {
	Address    chan *mail  // 邮箱地址
	AddressMap *addressMap // 通讯录
}

// 写邮件
func (this *mailBox) Write() draft {
	return draft{}
}

// 读邮件
func (this *mailBox) Read() mail {
	return mail{}
}

// 通讯录
type addressMap struct {
	lock       sync.Mutex
	addressMap map[chan *mailBox]*addressMap
}

// 添加好友
func (this *addressMap) AddFriend() bool {
	return true
}

// 删除好友
func (this *addressMap) DeleteFriend(mailBoxAddress chan *mailBox) {}

// 邮件
type mail struct {
	SenderMailBoxAddress chan *mail
	SenderName           string
	SendeeMailBoxAddress chan *mail
	SendeeName           string
	Content              interface{}
	Remarks              map[string]interface{}
}

// 回复
func (this mail) Reply() draft {
	return draft(this)
}

// 转发
func (this mail) Forward() draft {
	return draft(this)
}

// 草稿
type draft mail

// 发送
func (this draft) Send() bool {
	return true
}
