package como

import (
	"fmt"
	"sync"
	"time"
	"unsafe"
)

// 组织规划
type Provision struct {
	lock             sync.Mutex
	pactRegisterName string
	remainingTime    *time.Duration
	runningUpdates   map[chan interface{}]interface{}
	MailBox          mailBox
	T_T              leader
}

// 初始化(框架内部使用)
func (this *Provision) init(pactRegisterName string, mailLen int, remainingTime *time.Duration) (newMailBoxAddress chan mail) {
	this.pactRegisterName = pactRegisterName
	newMailBoxAddress = make(chan mail, mailLen)
	this.MailBox.Address = newMailBoxAddress
	this.MailBox.acceptLine = make(chan bool, 0)
	this.remainingTime = remainingTime
	return
}

// 投递邮件到邮箱(框架内部使用)
func (this *Provision) deliverMailForMailBox(newMail mail) {
	this.MailBox.mail = newMail
}

// 例行公事开始(框架内部使用)
func (this *Provision) routineStart() {}

// 例行公事收尾(框架内部使用)
func (this *Provision) routineEnd() {}

// 初始化组织
func (this *Provision) Init() {}

// 例行公事开始
func (this *Provision) RoutineStart() {}

// 例行公事收尾
func (this *Provision) RoutineEnd() {}

// 带外消息
func (this *Provision) Info() {}

// 组织解散
func (this *Provision) Terminate() {}

// 组织领导
type leader struct{}

// 添加事物循环处理
func (this *leader) AddUpdate() { fmt.Println("领导AddUpdate") }

// 拒绝本次服务
func (this *leader) DenialService() {
	org := (*Provision)(unsafe.Pointer(&*this))
	org.MailBox.mail.acceptLine <- false
}

// 获取超时时间
func (this *leader) GetRemainingTime() {}

// 设置超时时间
func (this *leader) SetRemainingTime() {}

// 邮箱
type mailBox struct {
	Address    chan mail   // 邮箱地址
	AddressMap *addressMap // 通讯录
	mail       mail        // 邮件
	acceptLine chan bool   // 询问别的组织受理用专线
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
	SenderAddress chan mail
	SenderName    string
	SendeeAddress chan mail
	SendeeName    string
	Content       interface{}
	Remarks       map[string]interface{}
	acceptLine    chan bool
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

// 发送(如果发送的地址或者接收人不存在返回false)
func (this draft) Send() bool {
	return true
}
