package como

import (
	"sync"
	"time"
	"unsafe"
)

// 组织规划
type Provision struct {
	T_T           leader     // 领导
	MailBox       mailBox    // 邮箱
	groupName     string     // 组织 组名称
	orgName       string     // 公约注册名称
	startTime     int64      // 开始时间
	endTime       int64      // 结束时间
	updateEndTime chan int64 // 更新结束时间的通道

}

// 初始化(框架内部使用)
func (this *Provision) init(groupName, orgName string, mailLen int, overtime int64) (newMailBoxAddress MailBoxAddress) {
	this.groupName = groupName
	this.orgName = orgName
	this.MailBox.Address.address = make(chan Mail, mailLen)
	this.MailBox.Address.mutex = &sync.Mutex{}
	this.MailBox.org = this
	isShut := false
	this.MailBox.Address.isShut = &isShut
	this.MailBox.AddressMap.addressMap = make(map[string]map[MailBoxAddress]struct{})
	this.T_T.runningUpdates = make(map[Uid]struct{})
	this.T_T.updateNotify = make(chan *updateInfo, 0)
	this.startTime = time.Now().Unix()
	this.endTime = this.startTime + overtime
	this.updateEndTime = make(chan int64, 0)
	return this.MailBox.Address
}

// 投递邮件到邮箱(框架内部使用)
func (this *Provision) deliverMailForMailBox(newMail Mail) {
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

// 组织解散
func (this *Provision) Terminate() {}

// 组织领导
type leader struct {
	currentAcceptState bool             // 是否接受受理了当前请求
	updateNotify       chan *updateInfo // 通知组织运行循环的通道
	runningUpdates     map[Uid]struct{} // 当前运行的附属循环
}

// 是否同意本次请求
func (this *leader) isAccept() bool {
	return this.currentAcceptState
}

// 添加事物循环处理并立刻运行一次function
func (this *leader) AddUpdateImmediateRun(function func(), timestep time.Duration) (uid Uid) {
	function()
	return this.AddUpdate(function, timestep)
}

// 添加事物循环处理
func (this *leader) AddUpdate(function func(), timestep time.Duration) (uid Uid) {
	updateColseChan := make(chan struct{}, 0)
	updateInfo := &updateInfo{updateColseChan: updateColseChan, isColse: false, function: function}
	uid = Uid(updateInfo)
	this.runningUpdates[uid] = struct{}{}
	go func() {
		for {
			select {
			case <-updateColseChan:
				return
			case <-time.After(timestep):
				select {
				case <-updateColseChan:
					return
				case this.updateNotify <- updateInfo:
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
	(*uid).close()
	delete(this.runningUpdates, uid)
	return
}

// 清理所有子循环
func (this *leader) CleanUpdates() {
	for uid, _ := range this.runningUpdates {
		(*uid).close()
		delete(this.runningUpdates, uid)
	}
}

// 拒绝本次服务
func (this *leader) DenyService() {
	this.currentAcceptState = false
}

// 获取注册组名称
func (this *leader) GetPactGroupName() string {
	org := (*Provision)(unsafe.Pointer(this))
	return org.groupName
}

// 获取注册名称
func (this *leader) GetPactRegisterName() string {
	org := (*Provision)(unsafe.Pointer(this))
	return org.orgName
}

// 获取超时时间
func (this *leader) GetOvertime() int64 {
	org := (*Provision)(unsafe.Pointer(this))
	return org.endTime - time.Now().Unix()
}

// 设置超时时间
func (this *leader) SetOvertime(newOvertime int64) {
	org := (*Provision)(unsafe.Pointer(this))
	org.endTime = time.Now().Unix() + newOvertime
	org.updateEndTime <- org.endTime
}

// 获取开始时间
func (this *leader) GetStartTime() int64 {
	org := (*Provision)(unsafe.Pointer(this))
	return org.startTime
}

// 获取结束时间
func (this *leader) GetEndTime() int64 {
	org := (*Provision)(unsafe.Pointer(this))
	return org.endTime
}

// 设置结束时间
func (this *leader) SetEndTime(newEndTime int64) {
	org := (*Provision)(unsafe.Pointer(this))
	org.endTime = newEndTime
	org.updateEndTime <- org.endTime
}

// 解散组织
func (this *leader) Dissolve() {
	org := (*Provision)(unsafe.Pointer(this))
	org.MailBox.Address.shut()
	org.T_T.CleanUpdates()
}

func (this *leader) getUpdateEndTimeChan() chan int64 {
	org := (*Provision)(unsafe.Pointer(this))
	return org.updateEndTime
}

// 子循环标识符
type Uid *updateInfo

// 循环详情
type updateInfo struct {
	updateColseChan chan struct{}
	isColse         bool
	function        func()
}

// 运行循环
func (this *updateInfo) run() {
	if this.isColse == true {
		return
	}
	this.function()
}

// 关闭循环
func (this *updateInfo) close() {
	if this.isColse == true {
		return
	}
	this.isColse = false
	close(this.updateColseChan)
}

// 邮箱
type mailBox struct {
	AddressMap addressMap     // 通讯录
	mail       Mail           // 邮件
	Address    MailBoxAddress // 邮箱地址
	org        *Provision     // 当前邮箱的组织
}

// 写邮件
func (this *mailBox) Write() Draft {
	return Draft{senderAddress: this.Address,
		senderGroupName:  this.org.groupName,
		senderOrgName:    this.org.orgName,
		senderServerName: this.mail.recipientServerName}
}

// 读邮件
func (this *mailBox) Read() *Mail {
	this.mail.recipientGroupName = this.org.groupName
	this.mail.recipientOrgName = this.org.orgName
	return &this.mail
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type MailBoxAddress struct {
	mutex   *sync.Mutex
	isShut  *bool     // 是否邮箱地址已经被关闭
	address chan Mail // 邮箱地址
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
func (this *MailBoxAddress) send(mail Mail) bool {
	this.mutex.Lock()         // 加锁
	defer this.mutex.Unlock() // 函数结束自动解锁
	if *this.isShut == true { // 如果当前地址已经关闭返回发送失败
		return false
	}
	this.address <- mail // 否则发送数据
	return true          // 返回发送失败
}

// 邮箱是否关闭?
func (this *MailBoxAddress) IsShut() bool {
	this.mutex.Lock()         // 加锁
	defer this.mutex.Unlock() // 函数结束自动解锁
	return *this.isShut       // 返回邮箱状态
}

// 通讯录
type addressMap struct {
	addressMap map[string]map[MailBoxAddress]struct{}
}

// 添加好友
func (this *addressMap) AddFriend(friendName string, mailBoxAddress MailBoxAddress) {
	_, ok := this.addressMap[friendName]
	if !ok {
		this.addressMap[friendName] = make(map[MailBoxAddress]struct{})
	}
	this.addressMap[friendName][mailBoxAddress] = struct{}{}
}

// 用名字删除用户
func (this *addressMap) RemoveFriends(friendsName string) {
	_, ok := this.addressMap[friendsName]
	if !ok {
		return
	}
	delete(this.addressMap, friendsName)
}

// 删除用户
func (this *addressMap) RemoveFriend(mailBoxAddress MailBoxAddress) {
	for friendsName, mailBoxsAddress := range this.addressMap {
		_, ok := mailBoxsAddress[mailBoxAddress]
		if ok {
			if len(this.addressMap) == 1 {
				delete(this.addressMap, friendsName)
				return
			}
			delete(mailBoxsAddress, mailBoxAddress)
			return
		}
	}
}

// 带名字删除好友
func (this *addressMap) RemoveFriendByName(friendName string, mailBoxAddress MailBoxAddress) {
	mailBoxsAddress, ok := this.addressMap[friendName]
	if !ok {
		return
	}
	_, ok = mailBoxsAddress[mailBoxAddress]
	if !ok {
		return
	}
	if len(this.addressMap[friendName]) == 1 {
		delete(this.addressMap, friendName)
		return
	}
	delete(mailBoxsAddress, mailBoxAddress)
}

// 删除所有好友
func (this *addressMap) RemoveAllFriends() {
	this.addressMap = make(map[string]map[MailBoxAddress]struct{})
}

// 获取好友
func (this *addressMap) GetFriends(friendsName string) (friendsMailBoxAddr []MailBoxAddress, ok bool) {
	mailBoxsAddress, ok := this.addressMap[friendsName]
	if !ok {
		return
	}
	friendsMailBoxAddr = make([]MailBoxAddress, 0, len(mailBoxsAddress))
	for mailBoxAddress, _ := range mailBoxsAddress {
		friendsMailBoxAddr = append(friendsMailBoxAddr, mailBoxAddress)
	}

	return
}

// 检查好友名字
func (this *addressMap) GetFriendName(mailBoxAddress MailBoxAddress) (friendName string, ok bool) {
	for friendName, addresss := range this.addressMap {
		for address, _ := range addresss {
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
		for mailBoxAddress, _ := range mailBoxsAddress {
			friendsMailBoxAddr = append(friendsMailBoxAddr, mailBoxAddress)
		}
	}
	return
}

// 发送给制定名称的用户
func (this *addressMap) SendForFriends(friendsName string, recipientServerName string, remarks map[string]interface{}, contents ...interface{}) {
	mailBoxsAddress, ok := this.addressMap[friendsName]
	if !ok {
		return
	}
	mailBox := (*mailBox)(unsafe.Pointer(this))
	Draft := mailBox.Write()
	Draft.recipientServerName = recipientServerName
	Draft.remarks = remarks
	for mailBoxAddress, _ := range mailBoxsAddress {
		Draft.recipientAddress = mailBoxAddress
		Draft.Send()
	}

}

// 发送数据给所有好友
func (this *addressMap) SendForAllFriends(recipientServerName string, remarks map[string]interface{}, contents ...interface{}) {
	mailBox := (*mailBox)(unsafe.Pointer(this))
	Draft := mailBox.Write()
	Draft.recipientServerName = recipientServerName
	Draft.remarks = remarks
	for _, mailBoxsAddress := range this.addressMap {
		for mailBoxAddress, _ := range mailBoxsAddress {
			Draft.recipientAddress = mailBoxAddress
			Draft.Send()
		}
	}
}

// 邮件
type Mail struct {
	senderAddress       MailBoxAddress         // 发件人地址
	senderGroupName     string                 // 发送人组名
	senderOrgName       string                 // 发送人组织名
	senderServerName    string                 // 发件人服务名字
	recipientAddress    MailBoxAddress         // 收件人地址
	recipientGroupName  string                 // 收件人组名
	recipientOrgName    string                 // 收件人组织名
	recipientServerName string                 // 收件人服务名字
	contents            []interface{}          // 邮件内容
	remarks             map[string]interface{} // 邮件备注
}

// 获取发件人地址
func (this *Mail) SenderAddress() MailBoxAddress {
	return this.senderAddress
}

// 获取发件人组名
func (this *Mail) SenderGroupName() string {
	return this.senderGroupName
}

// 获取发件人组织名
func (this *Mail) SenderOrgName() string {
	return this.senderOrgName
}

// 获取发件人服务名
func (this *Mail) SenderServerName() string {
	return this.senderServerName
}

// 获取收件人服务名
func (this *Mail) RecipientServerName() string {
	return this.recipientServerName
}

// 获取邮件内容
func (this *Mail) Content() []interface{} {
	return this.contents
}

// 获取一个邮件注释
func (this *Mail) Remark(key string) (val interface{}, ok bool) {
	if this.remarks != nil {
		val, ok = this.remarks[key]
		return
	}
	return nil, false
}

// 获取邮件注释
func (this *Mail) Remarks() map[string]interface{} {
	return this.remarks
}

// 回复
func (this Mail) Reply() Draft {
	Draft := Draft(this)                             // 创建新草稿
	senderAddress := Draft.senderAddress             // 临时储存发送者
	Draft.senderAddress = Draft.recipientAddress     // 接收地址变成发送地址
	Draft.senderGroupName = Draft.recipientGroupName // 接受组名变发送组名
	Draft.senderOrgName = Draft.recipientOrgName     // 接受组织名变发送组织名
	recipientServerName := Draft.senderServerName
	Draft.senderServerName = Draft.recipientServerName // 接收人变成发送人
	Draft.recipientAddress = senderAddress             // 发送地址变成接收地址
	Draft.recipientServerName = recipientServerName    // 设置接收人
	Draft.contents = nil                               // 内容待设置
	Draft.remarks = nil                                // 注释待设置
	return Draft                                       // 返回这个草稿
}

// 转发
func (this Mail) Forward(recipientAddress MailBoxAddress, recipientServerName string) Draft {
	Draft := Draft(this)                               // 创建新草稿
	Draft.senderAddress = Draft.recipientAddress       // 接收地址变成发送地址
	Draft.senderGroupName = Draft.recipientGroupName   // 接受组名变发送组名
	Draft.senderOrgName = Draft.recipientOrgName       // 接受组织名变发送组织名
	Draft.senderServerName = Draft.recipientServerName // 接收人变成发送人
	Draft.recipientAddress = recipientAddress          // 设置接收地址
	Draft.recipientServerName = recipientServerName    // 设置接收人
	return Draft                                       // 返回这个草稿
}

// 草稿
type Draft Mail

// 设置收件人地址
func (this *Draft) RecipientAddress(mailBoxAddress MailBoxAddress) {
	this.recipientAddress = mailBoxAddress
}

// 设置收件人名字
func (this *Draft) RecipientServerName(name string) {
	this.recipientServerName = name
}

// 设置发件人名字
func (this *Draft) SendServerName(name string) {
	this.senderServerName = name
}

// 设置草稿内容
func (this *Draft) Content(contents ...interface{}) {
	this.contents = contents
}

// 设置草稿备注
func (this *Draft) Remarks(remarks map[string]interface{}) {
	this.remarks = remarks
}

// 添加草稿备注
func (this *Draft) AddRemark(key string, val interface{}) {
	if this.remarks == nil {
		this.remarks = make(map[string]interface{})
	}
	this.remarks[key] = val
}

// 发送草稿(如果发送的地址或者接收人不存在返回false)
func (this Draft) Send() (ok bool) {

	if this.recipientAddress.address == nil {
		return false
	}

	return this.recipientAddress.send(Mail(this))
}
