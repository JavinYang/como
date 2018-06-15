package como

import (
	"reflect"
	"strconv"
	"sync"
	"time"
	"unsafe"
)

// 公约实例
var Pact *pacts

// 初始化公约实例
func pactInit() {
	// 创建公约
	staticOrg := &staticPact{groups: make(map[string]map[string]MailBoxAddress)}
	dynamicOrg := &dynamicPact{groups: make(map[string]map[string]*dynamicOrgInfo)}
	Pact = &pacts{staticOrg, dynamicOrg}
}

// 所有公约
type pacts struct {
	Static  *staticPact
	Dynamic *dynamicPact
}

// 组织规定规范
type provision interface {
	init(pactGroupName, pactRegisterName string, mailLen int, overTime int64) (newMailBoxAddress MailBoxAddress)
	deliverMailForMailBox(newMail mail)
	getLeader() *leader
	Init(...interface{})
	routineStart()
	RoutineStart()
	RoutineEnd()
	isMourn() bool
	Mourn()
	Terminate()
}

// 静态组织公约
type staticPact struct {
	groups map[string]map[string]MailBoxAddress
}

// 公开加入静态组织
func (this *staticPact) Join(groupName, orgName string, org provision, mailLen int, initPars ...interface{}) {

	group, ok := this.groups[groupName]
	if !ok {
		group = make(map[string]MailBoxAddress)
		goto NEW_ORG
	}

	_, ok = group[orgName]
	if ok {
		panic("组" + groupName + "已经存在叫做" + orgName + "的静态组织")
		return
	}

NEW_ORG:
	var overtime int64 = -1
	newMailBoxAddress := org.init(groupName, orgName, mailLen, overtime)
	group[orgName] = newMailBoxAddress
	this.groups[groupName] = group
	orgReflect := reflect.ValueOf(org)
	orgInstanceInfo := (&orgInstanceInfo{}).Init(orgReflect)

	runNeverTimeout(newMailBoxAddress, orgInstanceInfo, 0, nil, initPars...)
}

// 查询静态组织邮箱地址
func (this *staticPact) FindMailBoxAddress(groupName, orgName string) (mailBoxAddress MailBoxAddress, ok bool) {
	orgsMailBoxAddress, ok := this.groups[groupName]
	if !ok {
		return
	}
	mailBoxAddress, ok = orgsMailBoxAddress[orgName]
	return
}

// 获取制定静态组信息
func (this *staticPact) GetGroupInfo(groupName string) (GroupsInfo string) {
	group, ok := this.groups[groupName]
	if !ok {
		return
	}
	GroupsInfo = groupName
	for name, _ := range group {
		GroupsInfo += "\n" + "    " + name
	}
	return
}

// 获取所有静态组信息
func (this *staticPact) GetAllGroupInfo() (GroupsInfo string) {
	GroupsInfo += "\n----------STATIC_PACT----------\n"
	for groupName, _ := range this.groups {
		GroupsInfo += this.GetGroupInfo(groupName) + "\n"
	}
	return
}

// 动态组织公约
type dynamicPact struct {
	groups map[string]map[string]*dynamicOrgInfo
}

// 动态组织信息
type dynamicOrgInfo struct {
	orgSize  uintptr
	mailLen  int
	overtime int64
	pool     sync.Pool
}

// 加入动态组织 overtime == -1 永不超时
func (this *dynamicPact) Join(groupName, orgName string, org provision, mailLen int, overtime int64) {

	group, ok := this.groups[groupName]
	if !ok {
		group = make(map[string]*dynamicOrgInfo)
		goto NEW_ORG
	}

	_, ok = group[orgName]
	if ok {
		panic("已经存在叫做" + orgName + "的动态组织")
	}

NEW_ORG:
	orgType := reflect.Indirect(reflect.ValueOf(org)).Type() // 必须是值类型
	orgSize := orgType.Size()
	pool := sync.Pool{
		New: func() interface{} {
			orgReflect := reflect.New(orgType) // New新的肯定是指针
			return (&orgInstanceInfo{}).Init(orgReflect)
		},
	}

	group[orgName] = &dynamicOrgInfo{orgSize, mailLen, overtime, pool}
	this.groups[groupName] = group
}

// 生成已经加入的动态组织
func (this *dynamicPact) New(groupName, orgName string, initPars ...interface{}) (mailBoxAddress MailBoxAddress, ok bool) {
	group, ok := this.groups[groupName]
	if !ok {
		return
	}

	dynamicOrgInfo, ok := group[orgName]
	if !ok {
		return
	}
	overtime := dynamicOrgInfo.overtime

	orgInstanceInfo := dynamicOrgInfo.pool.Get().(*orgInstanceInfo)
	newMailBoxAddress := orgInstanceInfo.org.init(groupName, orgName, dynamicOrgInfo.mailLen, overtime)
	if overtime == -1 {
		runNeverTimeout(newMailBoxAddress, orgInstanceInfo, dynamicOrgInfo.orgSize, &dynamicOrgInfo.pool, initPars...)
	} else {
		runWithTimeout(newMailBoxAddress, orgInstanceInfo, dynamicOrgInfo.orgSize, &dynamicOrgInfo.pool, initPars...)
	}

	return newMailBoxAddress, true
}

// 获取制定动态组信息
func (this *dynamicPact) GetGroupInfo(groupName string) (GroupsInfo string) {
	group, ok := this.groups[groupName]
	if !ok {
		return
	}
	GroupsInfo = groupName
	for name, info := range group {
		GroupsInfo += "\n" + "    " + name + "    " + strconv.FormatInt(info.overtime, 10)
	}
	return
}

// 获取所有动态组信息
func (this *dynamicPact) GetAllGroupInfo() (GroupsInfo string) {
	GroupsInfo += "\n----------DYNAMIC_PACT----------\n"
	for groupName, _ := range this.groups {
		GroupsInfo += this.GetGroupInfo(groupName) + "\n"
	}
	return
}

// 运行无超时
func runNeverTimeout(newMailBoxAddress MailBoxAddress, orgInstanceInfo *orgInstanceInfo, orgSize uintptr, orgPool *sync.Pool, initPars ...interface{}) {

	go func() {

		org := orgInstanceInfo.org
		methodsMap := orgInstanceInfo.methodsMap

		T_T := org.getLeader()

		org.Init(initPars...)
		for {
			select {
			case <-newMailBoxAddress.isShut:
				org.Terminate()
				if orgPool != nil {
					memsetZero(orgInstanceInfo.pointer, orgSize)
					orgPool.Put(orgInstanceInfo)
				}
				return
			case mail, _ := <-newMailBoxAddress.address:
				if mail.systemTag == mailSystemTag_deathNotice {
					org.deliverMailForMailBox(mail)
					if org.isMourn() {
						org.Mourn()
					}
					continue
				}
				method, ok := methodsMap[mail.recipientServerName]
				if !ok {
					continue
				}
				org.deliverMailForMailBox(mail)
				org.routineStart()
				org.RoutineStart()
				if !T_T.isAccept() {
					continue
				}
				method()
				org.RoutineEnd()
				draftPoll.Put(draft(mail))
			case updateInfo, _ := <-T_T.updateNotify:
				updateInfo.run()
			}
		}
	}()
}

// 运行有超时
func runWithTimeout(newMailBoxAddress MailBoxAddress, orgInstanceInfo *orgInstanceInfo, orgSize uintptr, orgPool *sync.Pool, initPars ...interface{}) {

	org := orgInstanceInfo.org
	methodsMap := orgInstanceInfo.methodsMap

	T_T := org.getLeader()

	updateEndTime := T_T.getUpdateEndTimeChan()

	go func() {
		endTime := T_T.GetEndTime()
		ok := false
		for {
			select {
			case endTime, ok = <-updateEndTime:
				if !ok {
					return
				}
			case <-time.After(time.Duration((endTime - time.Now().Unix()) * 1e9)):
				T_T.timeOut()
				return
			}
		}
	}()

	go func() {

		org.Init(initPars...)
		for {
			select {
			case <-newMailBoxAddress.isShut:
				close(updateEndTime)
				org.Terminate()
				if orgPool != nil {
					memsetZero(orgInstanceInfo.pointer, orgSize)
					orgPool.Put(orgInstanceInfo)
				}
				return
			case mail, _ := <-newMailBoxAddress.address:
				if mail.systemTag != mailSystemTag__ {
					if mail.systemTag == mailSystemTag_deathNotice {
						org.deliverMailForMailBox(mail)
						if org.isMourn() {
							org.Mourn()
						}
						continue
					} else if mail.systemTag == mailSystemTag_timeOut {
						T_T.Dissolve()
						continue
					}
				}
				method, ok := methodsMap[mail.recipientServerName]
				if !ok {
					continue
				}
				org.deliverMailForMailBox(mail)
				org.routineStart()
				org.RoutineStart()
				if !T_T.isAccept() {
					continue
				}
				method()
				org.RoutineEnd()
				draftPoll.Put(draft(mail))
			case updateInfo, _ := <-T_T.updateNotify:
				updateInfo.run()
			}
		}
	}()
}

// 组织实例后的信息
type orgInstanceInfo struct {
	org        provision
	pointer    uintptr
	methodsMap map[string]func()
}

// 初始化组织实例信息
func (this *orgInstanceInfo) Init(orgReflect reflect.Value) *orgInstanceInfo {
	numMethod := orgReflect.NumMethod()
	this.methodsMap = make(map[string]func())
	for i := 0; i < numMethod; i++ {
		methodName := orgReflect.Type().Method(i).Name
		switch methodName {
		case "Init":
		case "RoutineStart":
		case "RoutineEnd":
		case "Terminate":
		default:
			this.methodsMap[methodName] = orgReflect.Method(i).Interface().(func())
		}
	}
	this.org = orgReflect.Interface().(provision) // 必须是指针
	this.pointer = orgReflect.Elem().UnsafeAddr() // 获取指针指向的的值的地址
	return this
}

// 初始化内存
func memsetZero(pointer uintptr, size uintptr) {

	if size < 1 { // 必须有要写0的内存尺寸
		return
	}

	tailPointer := pointer + size
	int64Size := unsafe.Sizeof(int64(0))

	if size < int64Size { // 不用加锁
		// 从头循到尾循环
		for ; pointer < tailPointer; pointer++ {
			pData := (*byte)(unsafe.Pointer(pointer))
			*pData = 0
		}
	} else { // 如果初始化的内存>=8字节才加速
		tail := size % int64Size              // 剩下的尾巴长度
		buttocksPointer := tailPointer - tail // 屁股的位置 = 尾巴位置 - 尾巴长度
		// 循环到屁股
		for ; pointer < buttocksPointer; pointer += int64Size {
			pData := (*int64)(unsafe.Pointer(pointer))
			*pData = 0
		}
		// 如果有尾巴就从屁股开始循环尾巴
		for ; buttocksPointer < tailPointer; buttocksPointer++ {
			pData := (*byte)(unsafe.Pointer(buttocksPointer))
			*pData = 0
		}
	}
}
