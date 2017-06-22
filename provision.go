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
	this.MailBox.AddressMap.addressMap = make(map[string][]MailBoxAddress)
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

// 获取注册名称
func (this *leader) GetPactRegisterName() string {
	org := (*Provision)(unsafe.Pointer(this))
	return org.pactRegisterName
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
	for _, mailBoxsAddress := range org.MailBox.AddressMap.addressMap {
		for _, mailBoxAddress := range mailBoxsAddress {
			draft.recipientAddress = mailBoxAddress
			draft.recipientName = "Info"
			closeRemark := make(map[string]interface{})
			closeRemark["GoodBye"] = nil
			draft.remarks = closeRemark
			draft.Send()
		}
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
	defer this.mutex.Unlock() // 解锁
	if *this.isShut == true { // 如果当前邮箱已经关闭直接返回
		return
	}
	close(this.address) // 关闭邮箱
	*this.isShut = true // 否则设置当前邮箱已关闭
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
	addressMap map[string][]MailBoxAddress
}

// 添加好友
func (this *addressMap) AddFriend(friendName string, mailBoxAddress MailBoxAddress) {
	mailBoxsAddress, ok := this.addressMap[friendName]
	if !ok {
		mailBoxsAddress = make([]MailBoxAddress, 0, 1)
	}
	this.addressMap[friendName] = append(mailBoxsAddress, mailBoxAddress)
}

// 用名字删除用户
func (this *addressMap) RemoveFriendByName(friendName string) {
	_, ok := this.addressMap[friendName]
	if !ok {
		return
	}
	delete(this.addressMap, friendName)
}

// 删除好友
func (this *addressMap) RemoveFriend(mailBoxAddress MailBoxAddress) {
	for friendName, mailBoxsAddress := range this.addressMap {
		mailBoxsAddressLen := len(mailBoxsAddress)
		for i := 0; i < mailBoxsAddressLen; i++ {
			if mailBoxsAddress[i] == mailBoxAddress {
				if mailBoxsAddressLen == 1 {
					delete(this.addressMap, friendName)
					return
				}
				mailBoxsAddress = append(mailBoxsAddress[:i], mailBoxsAddress[i+1:]...)
			}
		}
	}
}

// 删除所有好友
func (this *addressMap) RemoveAllFriends() {
	this.addressMap = make(map[string][]MailBoxAddress)
}

// 获取好友
func (this *addressMap) GetFriend(friendName string) (friendsMailBoxAddr []MailBoxAddress, ok bool) {
	friendsMailBoxAddr, ok = this.addressMap[friendName]
	return
}

// 检查好友名字
func (this *addressMap) GetFriendName(mailBoxAddress MailBoxAddress) (friendName string, ok bool) {
	for friendName, addresss := range this.addressMap {
		for _, address := range addresss {
			if address == mailBoxAddress {
				return friendName, true
			}
		}

	}
	return "", false
}

// 获取所有好友
func (this *addressMap) GetAllFriends() (friendsMailBoxAddr []MailBoxAddress) {
	friendsMailBoxAddr = make([]MailBoxAddress, 0, 10)
	for _, mailBoxsAddress := range this.addressMap {
		for _, mailBoxAddress := range mailBoxsAddress {
			friendsMailBoxAddr = append(friendsMailBoxAddr, mailBoxAddress)
		}
	}
	return
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

// 获取一个邮件注释
func (this *mail) GetRemark(key string) (val interface{}, ok bool) {
	if this.remarks != nil {
		val, ok = this.remarks[key]
		return
	}
	return nil, false
}

// 获取邮件注释
func (this *mail) GetRemarks() map[string]interface{} {
	return this.remarks
}

// 回复
func (this mail) Reply(recipientName string) draft {
	draft := draft(this)                         // 创建新草稿
	senderAddress := draft.senderAddress         // 临时储存发送者
	draft.senderAddress = draft.recipientAddress // 接收地址变成发送地址
	draft.senderName = draft.recipientName       // 接收人变成发送人
	draft.recipientAddress = senderAddress       // 发送地址变成接收地址
	draft.recipientName = recipientName          // 设置接收人
	draft.content = nil                          // 内容待设置
	draft.remarks = nil                          // 注释待设置
	return draft                                 // 返回这个草稿
}

// 转发
func (this mail) Forward(recipientAddress MailBoxAddress, recipientName string) draft {
	draft := draft(this)                         // 创建新草稿
	draft.senderAddress = draft.recipientAddress // 接收地址变成发送地址
	draft.senderName = draft.recipientName       // 接收人变成发送人
	draft.recipientAddress = recipientAddress    // 设置接收地址
	draft.recipientName = recipientName          // 设置接收人
	return draft                                 // 返回这个草稿
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

// 添加草稿备注
func (this *draft) AddRemark(key string, val interface{}) {
	if this.remarks == nil {
		this.remarks = make(map[string]interface{})
	}
	this.remarks[key] = val
}

// 发送草稿(如果发送的地址或者接收人不存在返回false)
func (this draft) Send() (ok bool) {

	if this.recipientAddress.address == nil {
		return false
	}

	return this.recipientAddress.send(mail(this))
}
