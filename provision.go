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
	this.MailBox.Address.mutex = &sync.Mutex{}
	this.MailBox.Address.isShut = make(chan struct{}, 0)
	this.MailBox.AddressMap.addressMap = make(map[string]map[MailBoxAddress]struct{})
	this.MailBox.AddressMap.deathNote = &sync.Map{}
	this.MailBox.AddressMap.mournNote = &sync.Map{}
	this.MailBox.authorizations = make(map[string]interface{})
	this.T_T.runningUpdates = sync.Map{}
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

// 判断是否调用吊念
func (this *Provision) isMourn() bool {
	val, ok := this.MailBox.AddressMap.mournNote.Load(this.MailBox.mail.senderAddress)
	if ok {
		this.MailBox.AddressMap.mournNote.Delete(this.MailBox.mail.senderAddress)
		if val.(bool) == true {
			this.T_T.Dissolve()
			return false
		}
		return true
	}
	return false
}

// 遗嘱
func (this *Provision) Mourn() {}

// 组织解散
func (this *Provision) Terminate() {}

// 组织领导
type leader struct {
	currentAcceptState bool             // 是否接受受理了当前请求
	updateNotify       chan *updateInfo // 通知组织运行循环的通道
	runningUpdates     sync.Map         // 当前运行的附属循环
}

// 是否同意本次请求
func (this *leader) isAccept() bool {
	return this.currentAcceptState
}

// 添加事物循环处理并立刻运行一次function
func (this *leader) AddUpdateImmediateRun(function func(), timestep time.Duration, times int64) (uid Uid) {
	if times != 0 {
		function()
		times -= 1
	}
	return this.AddUpdate(function, timestep, times)
}

// 添加事物循环处理
func (this *leader) AddUpdate(function func(), timestep time.Duration, times int64) (uid Uid) {
	updateColseChan := make(chan struct{}, 0)
	updateInfo := &updateInfo{updateColseChan: updateColseChan, isColse: false, function: function}
	uid = Uid(updateInfo)
	this.runningUpdates.Store(uid, struct{}{})
	go func() {
		if times < 0 { // 无限次循环
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
		} else if times > 0 { // 有限次循环
			var i int64 = 0
			for ; i < times; i++ {
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
			this.runningUpdates.Delete(uid)
		}
	}()
	return
}

// 清除事物循环处理
func (this *leader) RemoveUpdate(uid Uid) {
	_, ok := this.runningUpdates.Load(uid)
	if !ok {
		return
	}
	(*uid).close()
	this.runningUpdates.Delete(uid)
	return
}

// 清理所有子循环
func (this *leader) CleanUpdates() {
	this.runningUpdates.Range(func(key interface{}, value interface{}) bool {
		uid := key.(Uid)
		(*uid).close()
		this.runningUpdates.Delete(uid)
		return true
	})
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
func (this *leader) LinkDD(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你就死
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, true)

	// 你死我就死
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, true)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了我也死
		this.Dissolve()
	default:
	}
}

// 建立连接(我死你办事，你死我办事)
func (this *leader) LinkMM(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你办事
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, false)

	// 你死我办事
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, false)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了调用吊念函数
		this.sendTestamentForMe(org, hisAddr)
	default:
	}
}

// 建立连接(我死你也死，你死我办事)
func (this *leader) LinkDM(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你就死
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, true)

	// 你死我办事
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, false)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了调用吊念函数
		this.sendTestamentForMe(org, hisAddr)
	default:
	}
}

// 建立连接(我死你办事，你死我也死)
func (this *leader) LinkMD(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你办事
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, false)

	// 你死我就死
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, true)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了我也死
		this.Dissolve()
	default:
	}
}

// 建立连接(我死你也死，你死无所谓)
func (this *leader) LinkD_(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你就死
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, true)

	// 你死无所谓
	hisAddrMap.deathNote.Delete(myAddr)
	myAddrMap.mournNote.Delete(hisAddr)
}

// 建立连接(我死无所谓，你死我也死)
func (this *leader) Link_D(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死无所谓
	myAddrMap.deathNote.Delete(hisAddr)
	hisAddrMap.mournNote.Delete(myAddr)

	// 你死我就死
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, true)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了我也死
		this.Dissolve()
	default:
	}
}

// 建立连接(我死你办事，你死无所谓)
func (this *leader) LinkM_(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死你办事
	myAddrMap.deathNote.Store(hisAddr, struct{}{})
	hisAddrMap.mournNote.Store(myAddr, false)

	// 你死无所谓
	hisAddrMap.deathNote.Delete(myAddr)
	myAddrMap.mournNote.Delete(hisAddr)
}

// 建立连接(我死无所谓，你死我办事)
func (this *leader) Link_M(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死无所谓
	myAddrMap.deathNote.Delete(hisAddr)
	hisAddrMap.mournNote.Delete(myAddr)

	// 你死我办事
	hisAddrMap.deathNote.Store(myAddr, struct{}{})
	myAddrMap.mournNote.Store(hisAddr, false)

	select {
	case <-hisAddr.isShut: // 如果对面已经死了调用吊念函数
		this.sendTestamentForMe(org, hisAddr)
	default:
	}

}

// 解除连接
func (this *leader) Link__(hisAddr MailBoxAddress) {
	org := (*Provision)(unsafe.Pointer(this))
	if hisAddr == org.MailBox.Address || hisAddr.mailBox == nil {
		return
	}
	myAddr := org.MailBox.Address
	myAddrMap := org.MailBox.AddressMap
	hisAddrMap := hisAddr.mailBox.AddressMap

	// 我死无所谓
	myAddrMap.deathNote.Delete(hisAddr)
	hisAddrMap.mournNote.Delete(myAddr)

	// 你死无所谓
	hisAddrMap.deathNote.Delete(myAddr)
	myAddrMap.mournNote.Delete(hisAddr)
}

// 冒充对方发送遗嘱给自己
func (this *leader) sendTestamentForMe(org *Provision, mailBoxAddress MailBoxAddress) {
	draft := org.MailBox.Write()
	draft.systemTag = mailSystemTag_deathNotice
	draft.senderAddress = mailBoxAddress
	draft.senderServerName = ""
	draft.recipientAddress = org.MailBox.Address
	draft.Send()
}

// 更新时间用goroutine通知组织解体
func (this *leader) timeOut() {
	org := (*Provision)(unsafe.Pointer(this))
	draft := org.MailBox.Write()
	draft.systemTag = mailSystemTag_timeOut
	draft.senderServerName = ""
	draft.recipientAddress = org.MailBox.Address
	draft.Send()
}

// 解散组织
func (this *leader) Dissolve() {
	org := (*Provision)(unsafe.Pointer(this))
	org.T_T.CleanUpdates()     // 关闭循环
	org.MailBox.Address.shut() // 在关闭邮箱
}

// 获取剩余时间
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
	newDraft := draftPoll.Get().(draft)
	newDraft.systemTag = mailSystemTag__
	newDraft.senderAddress = this.Address
	newDraft.senderGroupName = this.org.groupName
	newDraft.senderOrgName = this.org.orgName
	newDraft.senderServerName = this.mail.recipientServerName
	return newDraft
}

// 读邮件
func (this *mailBox) Read() *mail {
	this.mail.recipientGroupName = this.org.groupName
	this.mail.recipientOrgName = this.org.orgName
	return &this.mail
}

// 邮箱地址(封装一层是因为不想让用户直接操作通道)
type MailBoxAddress struct {
	mutex   *sync.Mutex
	mailBox *mailBox      // 邮箱
	address chan mail     // 邮箱地址
	isShut  chan struct{} // 关闭用通道
}

// 关闭邮箱
func (this *MailBoxAddress) shut() {
	select {
	case <-this.isShut: // 如果已经关闭了不做操作
		return
	default:
		i := 0
		this.mailBox.AddressMap.deathNote.Range(
			func(key interface{}, val interface{}) bool {
				i++
				return true
			})

		// 关闭邮箱
		close(this.isShut)

		// 发送死亡通知
		this.mailBox.AddressMap.deathNote.Range(
			func(key interface{}, val interface{}) bool {
				draft := this.mailBox.Write()
				draft.systemTag = mailSystemTag_deathNotice
				draft.recipientAddress = key.(MailBoxAddress)
				draft.Send()
				return true
			})
	}
}

// 把数据放入邮箱
func (this *MailBoxAddress) send(mail mail) (ok bool) {

	select {
	case <-this.isShut:
		ok = false
	case this.address <- mail: // 否则发送数据
		ok = true // 返回发送失败
	}
	return
}

// 获取授权
func (this *MailBoxAddress) Remark(key string) (val interface{}, ok bool) {
	this.mutex.Lock()
	val, ok = this.mailBox.authorizations[key]
	this.mutex.Unlock()
	return
}

// 授权
func (this *MailBoxAddress) SetRemark(key string, value interface{}) {
	this.mutex.Lock()
	this.mailBox.authorizations[key] = value
	this.mutex.Unlock()
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
	addressMap map[string]map[MailBoxAddress]struct{}
	deathNote  *sync.Map // 我死的时候要通知的人
	mournNote  *sync.Map // 听这些人的遗嘱(ture表示我一起死，flase表示吊念)
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

// 邮箱系统标签枚举类型
type mailSystemTag int32

// 邮箱标签枚举
const (
	mailSystemTag__ mailSystemTag = iota
	mailSystemTag_deathNotice
	mailSystemTag_timeOut
)

// 邮件
type mail struct {
	systemTag           mailSystemTag  // 邮箱系统标识符
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
	return this.recipientAddress.send(mail(this))
}

// 草稿池
var draftPoll = sync.Pool{
	New: func() interface{} { return draft{} },
}
