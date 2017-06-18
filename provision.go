package como

import (
	"sync"
	"time"
	"unsafe"
)

// 组织规划
type Provision struct {
	T_T              leader  // 领导
	MailBox          mailBox // 邮箱
	pactRegisterName string  // 公约注册名称
	overtime         *int64  // 超时时间
}

// 初始化(框架内部使用)
func (this *Provision) init(pactRegisterName string, mailLen int, overtime *int64) (newMailBoxAddress MailBoxAddress) {
	this.pactRegisterName = pactRegisterName
	this.MailBox.Address.address = make(chan mail, mailLen)
	this.MailBox.Address.mutex = &sync.Mutex{}
	isShut := false
	this.MailBox.Address.isShut = &isShut
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
func (this *leader) GetOvertime() int64 {
	org := (*Provision)(unsafe.Pointer(this))
	if org.overtime == nil {
		return -1
	}
	return *org.overtime
}

// 设置超时时间
func (this *leader) SetOvertime(newOvertime int64) {
	org := (*Provision)(unsafe.Pointer(this))
	if org.overtime != nil {
		if newOvertime < 0 {
			newOvertime = 0
		}
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
		draft.recipientAddress = mailBoxAddress
		draft.recipientName = "Info"
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
}

// 写邮件
func (this *mailBox) Write() draft {
	return draft{senderAddress: this.Address, senderName: this.mail.recipientName}
}

// 读邮件
func (this *mailBox) Read() mail {
	return this.mail
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type MailBoxAddress struct {
	mutex   *sync.Mutex
	isShut  *bool     // 是否邮箱地址已经被关闭
	address chan mail // 邮箱地址
}

// 关闭邮箱
func (this *MailBoxAddress) shut() {
	this.mutex.Lock()         // 加锁
	if *this.isShut == true { // 如果当前邮箱已经关闭直接返回
		return
	}
	close(this.address) // 关闭邮箱
	*this.isShut = true // 否则设置当前邮箱已关闭
	this.mutex.Unlock() // 解锁
}

// 把数据放入邮箱
func (this *MailBoxAddress) send(mail mail) (isShut bool) {
	this.mutex.Lock()         // 加锁
	defer this.mutex.Unlock() // 函数结束自动解锁
	if *this.isShut == true { // 如果当前函数已经关闭返回关闭
		return true
	}
	this.address <- mail // 否则发送数据
	return false         // 返回没有关闭
}

// 邮箱是否关闭?
func (this *MailBoxAddress) IsShut() bool {
	this.mutex.Lock()         // 加锁
	defer this.mutex.Unlock() // 函数结束自动解锁
	return *this.isShut       // 返回邮箱状态
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
	senderAddress    MailBoxAddress         // 发件人地址
	senderName       string                 // 发件人名字
	recipientAddress MailBoxAddress         // 收件人地址
	recipientName    string                 // 收件人名字
	content          interface{}            // 邮件内容
	remarks          map[string]interface{} // 邮件备注
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
func (this mail) Reply(recipientName string) draft {
	draft := draft(this)
	senderAddress := draft.senderAddress
	draft.senderAddress = draft.recipientAddress
	draft.senderName = draft.recipientName
	draft.recipientAddress = senderAddress
	draft.recipientName = recipientName
	draft.content = nil
	draft.remarks = nil
	return draft
}

// 转发
func (this mail) Forward() draft {
	draft := draft(this)
	draft.senderAddress = draft.recipientAddress
	draft.senderName = draft.senderName
	draft.recipientAddress = MailBoxAddress{}
	draft.recipientName = ""
	return draft
}

// 草稿
type draft mail

// 设置收件人地址
func (this *draft) SetRecipientAddress(mailBoxAddress MailBoxAddress) {
	this.recipientAddress = mailBoxAddress
}

// 设置收件人名字
func (this *draft) SetRecipientName(sendeeName string) {
	this.recipientName = sendeeName
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

	if this.recipientAddress.address == nil {
		return false
	}

	if this.recipientAddress.address == this.senderAddress.address {
		panic("自己不能给自己发邮件!")
	}

	return this.recipientAddress.send(mail(this))
}
