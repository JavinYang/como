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
	this.MailBox.Address.address = newMailBoxAddress
	this.MailBox.acceptLine = make(chan bool, 0)
	this.remainingTime = remainingTime
	return
}

// 投递邮件到邮箱(框架内部使用)
func (this *Provision) deliverMailForMailBox(newMail mail) {
	this.MailBox.mail = newMail
}

// 是否正式受理请求
func (this *Provision) getLeader() leader {
	return this.T_T
}

// 例行公事开始(框架内部使用)
func (this *Provision) routineStart() {
	this.T_T.currentAcceptState = true
}

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
type leader struct {
	currentAcceptState bool
}

// 是否同意本次请求
func (this *leader) isAccept() bool {
	return this.currentAcceptState
}

// 添加事物循环处理
func (this *leader) AddUpdate() { fmt.Println("领导AddUpdate") }

// 拒绝本次服务
func (this *leader) DenialService() {
	this.currentAcceptState = false
}

// 获取超时时间
func (this *leader) GetRemainingTime() {}

// 设置超时时间
func (this *leader) SetRemainingTime() {}

// 解散组织
func (this *leader) Dissolve() {
	org := (*Provision)(unsafe.Pointer(&*this))
	close(org.MailBox.Address.address)
}

// 邮箱
type mailBox struct {
	Address    mailBoxAddress // 邮箱地址
	AddressMap *addressMap    // 通讯录
	mail       mail           // 邮件
	acceptLine chan bool      // 询问别的组织受理用专线
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type mailBoxAddress struct {
	address chan mail // 邮箱地址
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
	SenderAddress mailBoxAddress
	SenderName    string
	SendeeAddress mailBoxAddress
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
func (this draft) Send() (ok bool) {

	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	this.SendeeAddress.address <- mail(this)
	ok = <-this.acceptLine

	return
}
