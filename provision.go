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
	this.MailBox.org = this
	this.MailBox.Address.mailBox = &this.MailBox
	this.MailBox.Address.address = make(chan mail, mailLen)
	this.MailBox.Address.mutex = &sync.RWMutex{}
	isShut := false
	this.MailBox.Address.isShut = &isShut
	this.MailBox.AddressMap.addressMap = make(map[string]map[MailBoxAddress]struct{})
	this.MailBox.AddressMap.dissolveAddressMap = make(map[MailBoxAddress]struct{})
	this.MailBox.AddressMap.testamentAddressMap = make(map[MailBoxAddress]struct{})
	this.MailBox.authorizations = make(map[string]interface{})
	this.T_T.runningUpdates = make(map[Uid]struct{})
	this.T_T.updateNotify = make(chan *updateInfo, 0)
	this.startTime = time.Now().Unix()
	this.endTime = this.startTime + overtime
	this.updateEndTime = make(chan int64, 0)
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

// 遗嘱
func (this *Provision) Testament() {}

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

// 建立连接(共生死)
func (this *leader) LinkDD(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了那我也要死
		mailBoxAddress.mutex.Unlock()
		this.Dissolve()
	} else {
		// 我死你也死
		mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 你死我也死
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 建立连接(我死告诉你，你死告诉我)
func (this *leader) LinkTT(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了调用遗嘱函数
		mailBoxAddress.mutex.Unlock()
		this.sendTestamentForMe(org, mailBoxAddress)
	} else {
		// 你死告诉我
		mailBoxAddress.mailBox.AddressMap.testamentAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 我死告诉你
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.testamentAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 建立连接(我死你也死，你死告诉我)
func (this *leader) LinkDT(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了调用遗嘱函数
		mailBoxAddress.mutex.Unlock()
		this.sendTestamentForMe(org, mailBoxAddress)
	} else {
		// 你死告诉我
		mailBoxAddress.mailBox.AddressMap.testamentAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 我死你也死
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}

}

// 建立连接(我死告诉你，你死我也死)
func (this *leader) LinkTD(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了调用遗嘱函数
		mailBoxAddress.mutex.Unlock()
		this.sendTestamentForMe(org, mailBoxAddress)
	} else {
		// 你死我也死
		mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 我死告诉你
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.testamentAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 建立连接(我死你也死，你死无所谓)
func (this *leader) LinkD_(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	// 如果你没死
	if !*mailBoxAddress.isShut {
		// 你死我所谓
		_, ok := mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address]
		if ok {
			delete(mailBoxAddress.mailBox.AddressMap.dissolveAddressMap, org.MailBox.Address)
		}
		mailBoxAddress.mutex.Unlock()
		// 我死你也死
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}

}

// 建立连接(我死无所谓，你死我也死)
func (this *leader) Link_D(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	org.MailBox.Address.mutex.Unlock()
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了我也死
		mailBoxAddress.mutex.Unlock()
		this.Dissolve()
	} else {
		// 你死我就死
		mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 我死无所谓
		org.MailBox.Address.mutex.Lock()
		_, ok := org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress]
		if ok {
			delete(org.MailBox.AddressMap.dissolveAddressMap, mailBoxAddress)
		}
	}

}

// 建立连接(我死告诉你，你死无所谓)
func (this *leader) LinkT_(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if !*mailBoxAddress.isShut {
		// 你死无所谓
		_, ok := mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address]
		if ok {
			delete(mailBoxAddress.mailBox.AddressMap.dissolveAddressMap, org.MailBox.Address)
		}
		mailBoxAddress.mutex.Unlock()
		// 我死告诉你
		org.MailBox.Address.mutex.Lock()
		org.MailBox.AddressMap.testamentAddressMap[mailBoxAddress] = struct{}{}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 建立连接(我死无所谓，你死告诉我)
func (this *leader) Link_T(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if *mailBoxAddress.isShut { // 如果对面已经死了冒充对方发送遗嘱给自己
		mailBoxAddress.mutex.Unlock()
		this.sendTestamentForMe(org, mailBoxAddress)
	} else {
		// 你死告诉我
		mailBoxAddress.mailBox.AddressMap.testamentAddressMap[org.MailBox.Address] = struct{}{}
		mailBoxAddress.mutex.Unlock()
		// 我死无所谓
		org.MailBox.Address.mutex.Lock()
		_, ok := org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress]
		if ok {
			delete(org.MailBox.AddressMap.dissolveAddressMap, mailBoxAddress)
		}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 解除连接
func (this *leader) Link__(mailBoxAddress MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if mailBoxAddress == org.MailBox.Address || mailBoxAddress.mailBox == nil {
		return
	}
	mailBoxAddress.mutex.Lock()
	if !*mailBoxAddress.isShut {
		// 你死无所谓
		_, ok := mailBoxAddress.mailBox.AddressMap.dissolveAddressMap[org.MailBox.Address]
		if ok {
			delete(mailBoxAddress.mailBox.AddressMap.dissolveAddressMap, org.MailBox.Address)
		}
		mailBoxAddress.mutex.Unlock()
		// 我死无所谓
		org.MailBox.Address.mutex.Lock()
		_, ok = org.MailBox.AddressMap.dissolveAddressMap[mailBoxAddress]
		if ok {
			delete(org.MailBox.AddressMap.dissolveAddressMap, mailBoxAddress)
		}
		org.MailBox.Address.mutex.Unlock()
	}
}

// 冒充对方发送遗嘱给自己
func (this *leader) sendTestamentForMe(org *Provision, mailBoxAddress MailBoxAddress) {
	draft := org.MailBox.Write()
	draft.senderAddress = mailBoxAddress
	draft.senderServerName = ""
	draft.recipientAddress = org.MailBox.Address
	draft.recipientServerName = "Testament"
	draft.Send()
}

// 解散组织
func (this *leader) Dissolve() {
	org := (*Provision)(unsafe.Pointer(this))
	org.T_T.CleanUpdates()     // 关闭循环
	org.MailBox.Address.shut() // 在关闭邮箱
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
	AddressMap     addressMap             // 通讯录
	mail           mail                   // 邮件
	Address        MailBoxAddress         // 邮箱地址
	org            *Provision             // 当前邮箱的组织
	authorizations map[string]interface{} // 授权
}

// 写邮件
func (this *mailBox) Write() draft {
	return draft{isSystem: false,
		senderAddress:    this.Address,
		senderGroupName:  this.org.groupName,
		senderOrgName:    this.org.orgName,
		senderServerName: this.mail.recipientServerName}
}

// 读邮件
func (this *mailBox) Read() *mail {
	this.mail.recipientGroupName = this.org.groupName
	this.mail.recipientOrgName = this.org.orgName
	return &this.mail
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type MailBoxAddress struct {
	mutex   *sync.RWMutex
	mailBox *mailBox  // 邮箱
	isShut  *bool     // 是否邮箱地址已经被关闭
	address chan mail // 邮箱地址
}

// 关闭邮箱
func (this *MailBoxAddress) shut() {
	this.mutex.Lock()         // 加锁
	if *this.isShut == true { // 如果当前邮箱已经关闭直接返回
		this.mutex.Unlock() // 解锁
		return
	}
	close(this.address) // 关闭邮箱
	// 发送解散通知
	for mailBoxAddress, _ := range this.mailBox.AddressMap.dissolveAddressMap {
		this.sendDissolve(mailBoxAddress)
	}
	// 发送遗嘱
	for mailBoxAddress, _ := range this.mailBox.AddressMap.testamentAddressMap {
		this.sendTestament(mailBoxAddress)
	}
	*this.isShut = true // 否则设置当前邮箱已关闭
	this.mutex.Unlock() // 解锁
}

// 发送遗嘱
func (this *MailBoxAddress) sendTestament(mailBoxAddress MailBoxAddress) {
	draft := this.mailBox.Write()
	draft.isSystem = true
	draft.recipientAddress = mailBoxAddress
	draft.recipientServerName = "Testament"
	draft.Send()
}

// 发送解散
func (this *MailBoxAddress) sendDissolve(mailBoxAddress MailBoxAddress) {
	draft := this.mailBox.Write()
	draft.isSystem = true
	draft.recipientAddress = mailBoxAddress
	draft.recipientServerName = "dissolve"
	draft.Send()
}

// 把数据放入邮箱
func (this *MailBoxAddress) send(mail mail) (ok bool) {
	this.mutex.Lock() // 加锁
	if *this.isShut { // 如果当前地址已经关闭返回发送失败
		this.mutex.Unlock() // 函数结束自动解锁
		ok = false
		return
	}
	this.address <- mail // 否则发送数据
	ok = true            // 返回发送失败
	this.mutex.Unlock()  // 函数结束自动解锁
	return
}

// 授权
func (this *MailBoxAddress) Remark(key string) (val interface{}, ok bool) {
	this.mutex.RLock()
	val, ok = this.mailBox.authorizations[key]
	this.mutex.RUnlock()
	return
}

// 删除授权
func (this *MailBoxAddress) RemoveRemark(key string) {
	this.mutex.Lock()
	_, ok := this.mailBox.authorizations[key]
	if ok {
		delete(this.mailBox.authorizations, key)
	}
	this.mutex.Unlock()
}

// 删除所有授权
func (this *MailBoxAddress) RemoveAllRemarks() {
	authorizations := this.mailBox.authorizations
	this.mutex.Lock()
	for key, _ := range authorizations {
		delete(authorizations, key)
	}
	this.mutex.Unlock()
}

// 通讯录
type addressMap struct {
	addressMap          map[string]map[MailBoxAddress]struct{}
	dissolveAddressMap  map[MailBoxAddress]struct{} // 我死的时候要杀死的人
	testamentAddressMap map[MailBoxAddress]struct{} // 我死的时候要通知的人
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
	draft := mailBox.Write()
	draft.recipientServerName = recipientServerName
	for mailBoxAddress, _ := range mailBoxsAddress {
		draft.recipientAddress = mailBoxAddress
		draft.Send()
	}

}

// 发送数据给所有好友
func (this *addressMap) SendForAllFriends(recipientServerName string, remarks map[string]interface{}, contents ...interface{}) {
	mailBox := (*mailBox)(unsafe.Pointer(this))
	draft := mailBox.Write()
	draft.recipientServerName = recipientServerName
	for _, mailBoxsAddress := range this.addressMap {
		for mailBoxAddress, _ := range mailBoxsAddress {
			draft.recipientAddress = mailBoxAddress
			draft.Send()
		}
	}
}

// 邮件
type mail struct {
	isSystem            bool           // 是否是系统邮件
	senderAddress       MailBoxAddress // 发件人地址
	senderGroupName     string         // 发送人组名
	senderOrgName       string         // 发送人组织名
	senderServerName    string         // 发件人服务名字
	recipientAddress    MailBoxAddress // 收件人地址
	recipientGroupName  string         // 收件人组名
	recipientOrgName    string         // 收件人组织名
	recipientServerName string         // 收件人服务名字
	contents            []interface{}  // 邮件内容
}

// 获取发件人地址
func (this *mail) SenderAddress() MailBoxAddress {
	return this.senderAddress
}

// 获取发件人组名
func (this *mail) SenderGroupName() string {
	return this.senderGroupName
}

// 获取发件人组织名
func (this *mail) SenderOrgName() string {
	return this.senderOrgName
}

// 获取发件人服务名
func (this *mail) SenderServerName() string {
	return this.senderServerName
}

// 获取收件人服务名
func (this *mail) RecipientServerName() string {
	return this.recipientServerName
}

// 获取邮件内容
func (this *mail) Content() []interface{} {
	return this.contents
}

// 回复
func (this mail) Reply() draft {
	draft := draft(this)                               // 创建新草稿
	senderAddress := draft.senderAddress               // 临时储存发送者
	draft.senderAddress = draft.recipientAddress       // 接收地址变成发送地址
	draft.senderGroupName = draft.recipientGroupName   // 接受组名变发送组名
	draft.senderOrgName = draft.recipientOrgName       // 接受组织名变发送组织名
	recipientServerName := draft.senderServerName      // 接受人服务名称
	draft.senderServerName = draft.recipientServerName // 接收人变成发送人
	draft.recipientAddress = senderAddress             // 发送地址变成接收地址
	draft.recipientServerName = recipientServerName    // 设置接收人
	draft.contents = nil                               // 内容待设置
	return draft                                       // 返回这个草稿
}

// 转发
func (this mail) Forward(recipientAddress MailBoxAddress, recipientServerName string) draft {
	draft := draft(this)                               // 创建新草稿
	draft.senderAddress = draft.recipientAddress       // 接收地址变成发送地址
	draft.senderGroupName = draft.recipientGroupName   // 接受组名变发送组名
	draft.senderOrgName = draft.recipientOrgName       // 接受组织名变发送组织名
	draft.senderServerName = draft.recipientServerName // 接收人变成发送人
	draft.recipientAddress = recipientAddress          // 设置接收地址
	draft.recipientServerName = recipientServerName    // 设置接收人
	return draft                                       // 返回这个草稿
}

// 草稿
type draft mail

// 设置收件人地址
func (this *draft) RecipientAddress(mailBoxAddress MailBoxAddress) {
	this.recipientAddress = mailBoxAddress
}

// 设置收件人名字
func (this *draft) RecipientServerName(name string) {
	this.recipientServerName = name
}

// 设置发件人名字
func (this *draft) SendServerName(name string) {
	this.senderServerName = name
}

// 设置草稿内容
func (this *draft) Content(contents ...interface{}) {
	this.contents = contents
}

// 发送草稿(如果发送的地址或者接收人不存在返回false)
func (this draft) Send() (ok bool) {

	if this.recipientAddress.address == nil {
		return false
	}

	// 发送给自己
	if this.recipientAddress.address == this.senderAddress.address {
		this.recipientAddress.address <- mail(this)
		return true
	}

	return this.recipientAddress.send(mail(this))
}
