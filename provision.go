package como

import (
	"sync"
	"time"
	"unsafe"
)

// 组织规划
type Provision struct {
	T_T              leader         // 领导
	MailBox          mailBox        // 邮箱
	pactRegisterName string         // 公约注册名称
	overtime         *time.Duration // 超时时间
}

// 初始化(框架内部使用)
func (this *Provision) init(pactRegisterName string, mailLen int, overtime *time.Duration) (newMailBoxAddress MailBoxAddress) {
	this.pactRegisterName = pactRegisterName
	this.MailBox.Address.address = make(chan mail, mailLen)
	this.MailBox.Address.mutex = &sync.Mutex{}
	isShut := false
	this.MailBox.Address.isShut = &isShut
	this.MailBox.acceptLine = make(chan bool, 0)
	this.MailBox.AddressMap.addressMap = make(map[MailBoxAddress]*struct{})
	this.overtime = overtime
	this.T_T.runningUpdates = make(map[Uid]*struct{})
	this.T_T.updateNotify = make(chan func(), 0)
	return this.MailBox.Address
}

// 投递邮件到邮箱(框架内部使用)
func (this *Provision) deliverMailForMailBox(newMail mail) {
	this.MailBox.mail = newMail
}

// 获取领导
func (this *Provision) getLeader() *leader {
	return &this.T_T
}

// 例行公事开始(框架内部使用)
func (this *Provision) routineStart() {
	this.T_T.currentAcceptState = true
}

// 初始化组织
func (this *Provision) Init(pars ...interface{}) {}

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
	currentAcceptState bool              // 是否接受受理了当前请求
	updateNotify       chan func()       // 通知组织运行循环的通道
	runningUpdates     map[Uid]*struct{} // 当前运行的附属循环
}

// 是否同意本次请求
func (this *leader) isAccept() bool {
	return this.currentAcceptState
}

// 添加事物循环处理
func (this *leader) AddUpdate(function func(), timestep time.Duration) (uid Uid) {
	updateColseChan := make(chan struct{}, 0)
	uid = Uid{updateColseChan}
	this.runningUpdates[uid] = nil
	go func() {
		for {
			select {
			case <-updateColseChan:
				return
			case <-time.After(timestep):
				select {
				case <-updateColseChan:
					return
				case this.updateNotify <- function:
				}
			}
		}
	}()
	return
}

// 清除事物循环处理
func (this *leader) RemoveUpdate(uid Uid) {
	_, ok := this.runningUpdates[uid]
	if !ok {
		return
	}
	delete(this.runningUpdates, uid)
	close(uid.updateColseChan)
	return
}

// 清理所有子循环
func (this *leader) CleanUpdates() {
	for updateId, _ := range this.runningUpdates {
		delete(this.runningUpdates, updateId)
		close(updateId.updateColseChan)
	}
}

// 拒绝本次服务
func (this *leader) DenyService() {
	this.currentAcceptState = false
}

// 获取超时时间
func (this *leader) GetOvertime() time.Duration {
	org := (*Provision)(unsafe.Pointer(this))
	if org.overtime == nil {
		return -1
	}
	return *org.overtime
}

// 设置超时时间
func (this *leader) SetOvertime(newOvertime time.Duration) {
	org := (*Provision)(unsafe.Pointer(this))
	if org.overtime != nil {
		*org.overtime = newOvertime
	}
}

// 解散组织
func (this *leader) Dissolve() {
	org := (*Provision)(unsafe.Pointer(this))
	org.MailBox.Address.shut()
	org.T_T.CleanUpdates()
}

// 跟所有朋友告别(框架内部使用)
func (this *leader) goodByeMyFriends() {
	org := (*Provision)(unsafe.Pointer(this))
	draft := org.MailBox.Write()
	draft.senderName = "Info"
	for mailBoxAddress, _ := range org.MailBox.AddressMap.addressMap {
		draft.sendeeAddress = mailBoxAddress
		draft.sendeeName = "Info"
		closeRemark := make(map[string]interface{})
		closeRemark["GoodBye"] = nil
		draft.remarks = closeRemark
		draft.Send()
	}
}

// 子循环标识符
type Uid struct {
	updateColseChan chan struct{}
}

// 邮箱
type mailBox struct {
	mail       mail           // 邮件
	Address    MailBoxAddress // 邮箱地址
	AddressMap addressMap     // 通讯录
	acceptLine chan bool      // 询问别的组织受理用专线
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type MailBoxAddress struct {
	mutex   *sync.Mutex
	address chan mail // 邮箱地址
	isShut  *bool     // 是否邮箱地址已经被关闭
}

// 写邮件
func (this *mailBox) Write() draft {
	return draft{senderAddress: this.Address, senderName: this.mail.sendeeName, acceptLine: this.acceptLine}
}

// 读邮件
func (this *mailBox) Read() mail {
	return this.mail // 这里是否会读到同一封邮件?!!!
}

// 关闭邮箱
func (this *MailBoxAddress) shut() {
	this.mutex.Lock()
	if *this.isShut == true {
		return
	}
	close(this.address)
	*this.isShut = true
	this.mutex.Unlock()
}

// 把数据放入邮箱
func (this *MailBoxAddress) send(mail mail) (isShut bool) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	if *this.isShut == true {
		return true
	}
	this.address <- mail
	return false
}

// 邮箱是否关闭?
func (this *MailBoxAddress) IsShut() bool {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	return *this.isShut
}

// 通讯录
type addressMap struct {
	addressMap map[MailBoxAddress]*struct{}
}

// 添加好友
func (this *addressMap) AddFriend(mailBoxAddress MailBoxAddress) {
	this.addressMap[mailBoxAddress] = nil
}

// 删除好友
func (this *addressMap) RemoveFriend(mailBoxAddress MailBoxAddress) {
	_, ok := this.addressMap[mailBoxAddress]
	if !ok {
		return
	}
	delete(this.addressMap, mailBoxAddress)
	return
}

// 获取所有好友
func (this *addressMap) GetAllFriends() []MailBoxAddress {
	friendsAddress := make([]MailBoxAddress, len(this.addressMap))
	i := 0
	for address, _ := range this.addressMap {
		friendsAddress[i] = address
		i++
	}
	return friendsAddress
}

// 邮件
type mail struct {
	senderAddress MailBoxAddress         // 发件人地址
	senderName    string                 // 发件人名字
	sendeeAddress MailBoxAddress         // 收件人地址
	sendeeName    string                 // 收件人名字
	content       interface{}            // 邮件内容
	remarks       map[string]interface{} // 邮件备注
	acceptLine    chan bool              // 请求专线用来等待收件方受理请求
}

// 获取发件人地址
func (this *mail) GetSenderAddress() MailBoxAddress {
	return this.senderAddress
}

// 获取发件人名字
func (this *mail) GetSenderName() string {
	return this.senderName
}

// 获取邮件内容
func (this *mail) GetContent() interface{} {
	return this.content
}

// 获取邮件注释
func (this *mail) GetRemarks() map[string]interface{} {
	return this.remarks
}

// 回复
func (this mail) Reply(sendeeName string) draft {
	draft := draft(this)
	senderAddress := draft.senderAddress
	draft.senderAddress = draft.sendeeAddress
	draft.senderName = draft.sendeeName
	draft.sendeeAddress = senderAddress
	draft.sendeeName = sendeeName
	draft.content = nil
	draft.remarks = nil
	return draft
}

// 转发
func (this mail) Forward() draft {
	draft := draft(this)
	draft.senderAddress = draft.sendeeAddress
	draft.senderName = draft.senderName
	draft.sendeeAddress = MailBoxAddress{}
	draft.sendeeName = ""
	return draft
}

// 草稿
type draft mail

// 设置收件人地址
func (this *draft) SetSendeeAddress(mailBoxAddress MailBoxAddress) {
	this.sendeeAddress = mailBoxAddress
}

// 设置收件人名字
func (this *draft) SetSendeeName(sendeeName string) {
	this.sendeeName = sendeeName
}

// 设置草稿内容
func (this *draft) SetContent(content interface{}) {
	this.content = content
}

// 设置草稿备注
func (this *draft) SetRemarks(remarks map[string]interface{}) {
	this.remarks = remarks
}

// 发送草稿(如果发送的地址或者接收人不存在返回false)
func (this draft) Send() (ok bool) {

	if this.sendeeAddress.address == nil {
		return false
	}

	if this.sendeeAddress.address == this.senderAddress.address {
		panic("自己不能给自己发邮件!")
	}

	isShut := this.sendeeAddress.send(mail(this))
	if !isShut {
		ok = <-this.acceptLine
	}

	return
}
