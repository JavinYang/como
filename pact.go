package como

import (
	"reflect"
	"time"
)

// 公约实例
var Pact *pacts

// 初始化公约实例
func init() {
	staticOrg := &staticPact{groups: make(map[string]map[string]MailBoxAddress)}
	dynamicOrg := &dynamicPact{groups: make(map[string]map[string]dynamicOrgInfo)}
	Pact = &pacts{staticOrg, dynamicOrg}
}

// 所有公约
type pacts struct {
	Static  *staticPact
	Dynamic *dynamicPact
}

// 创建一个新动态组织带超时控制 overtime == -1 时永不超时
func (this *pacts) New(org provision, mailLen int, overtime int64, initPars ...interface{}) (mailBoxAddress MailBoxAddress, ok bool) {
	newMailBoxAddress := org.init("", "", mailLen, &overtime)
	orgReflect := reflect.ValueOf(org)
	if overtime == -1 {
		runNeverTimeout(newMailBoxAddress, orgReflect, org, initPars...)
	} else {
		runWithTimeout(newMailBoxAddress, orgReflect, org, overtime, initPars...)
	}

	return newMailBoxAddress, true
}

// 组织规定规范
type provision interface {
	init(pactGroupName, pactRegisterName string, mailLen int, overTime *int64) (newMailBoxAddress MailBoxAddress)
	deliverMailForMailBox(newMail mail)
	getLeader() *leader
	Init(...interface{})
	routineStart()
	RoutineStart()
	RoutineEnd()
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
	newMailBoxAddress := org.init(groupName, orgName, mailLen, &overtime)
	group[orgName] = newMailBoxAddress
	this.groups[groupName] = group
	orgReflect := reflect.ValueOf(org)

	runNeverTimeout(newMailBoxAddress, orgReflect, org, initPars...)
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
	GroupsInfo += "\n----------STATIC_PATH----------\n"
	for groupName, _ := range this.groups {
		GroupsInfo += this.GetGroupInfo(groupName) + "\n"
	}
	return
}

// 动态组织公约
type dynamicPact struct {
	groups map[string]map[string]dynamicOrgInfo
}

// 动态组织信息
type dynamicOrgInfo struct {
	orgType  reflect.Type
	mailLen  int
	overtime int64
}

// 加入动态组织 overtime == -1 永不超时
func (this *dynamicPact) Join(groupName, orgName string, provision provision, mailLen int, overtime int64) {

	group, ok := this.groups[groupName]
	if !ok {
		group = make(map[string]dynamicOrgInfo)
		goto NEW_ORG
	}

	_, ok = group[orgName]
	if ok {
		panic("已经存在叫做" + orgName + "的动态组织")
	}

NEW_ORG:
	group[orgName] = dynamicOrgInfo{reflect.Indirect(reflect.ValueOf(provision)).Type(), mailLen, overtime}
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
	orgReflect := reflect.New(dynamicOrgInfo.orgType)
	org := orgReflect.Interface().(provision)
	newMailBoxAddress := org.init(groupName, orgName, dynamicOrgInfo.mailLen, &overtime)
	if overtime == -1 {
		runNeverTimeout(newMailBoxAddress, orgReflect, org, initPars...)
	} else {
		runWithTimeout(newMailBoxAddress, orgReflect, org, overtime, initPars...)
	}

	return newMailBoxAddress, true
}

// 获取制定静态组信息
func (this *dynamicPact) GetGroupInfo(groupName string) (GroupsInfo string) {
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
func (this *dynamicPact) GetAllGroupInfo() (GroupsInfo string) {
	GroupsInfo += "\n----------Dynamic_Pact----------\n"
	for groupName, _ := range this.groups {
		GroupsInfo += this.GetGroupInfo(groupName) + "\n"
	}
	return
}

// 运行动态组织
func runWithTimeout(newMailBoxAddress MailBoxAddress, orgReflect reflect.Value, org provision, overtime int64, initPars ...interface{}) {

	planningMethodsMap := make(map[string]func())
	numMethod := orgReflect.NumMethod()
	for i := 0; i < numMethod; i++ {
		methodName := orgReflect.Type().Method(i).Name
		switch methodName {
		case "Init":
		case "RoutineStart":
		case "RoutineEnd":
		case "Terminate":
		default:
			planningMethodsMap[methodName] = orgReflect.Method(i).Interface().(func())
		}
	}

	T_T := org.getLeader()

	go func() {
		org.Init(initPars...)
		for {
			select {
			case mail, ok := <-newMailBoxAddress.address:
				if !ok {
					org.Terminate()
					return
				}

				method, ok := planningMethodsMap[mail.recipientServerName]
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
			case function, _ := <-T_T.updateNotify:
				function()
			case <-time.After(time.Second):
				if overtime == 0 {
					T_T.Dissolve()
				}
				overtime -= 1
			}
		}
	}()
}

// 运行静态组织
func runNeverTimeout(newMailBoxAddress MailBoxAddress, orgReflect reflect.Value, org provision, initPars ...interface{}) {

	planningMethodsMap := make(map[string]func())
	numMethod := orgReflect.NumMethod()
	for i := 0; i < numMethod; i++ {
		methodName := orgReflect.Type().Method(i).Name
		switch methodName {
		case "Init":
		case "RoutineStart":
		case "RoutineEnd":
		case "Terminate":
		default:
			planningMethodsMap[methodName] = orgReflect.Method(i).Interface().(func())
		}
	}

	T_T := org.getLeader()
	go func() {
		org.Init(initPars...)
		for {
			select {
			case mail, ok := <-newMailBoxAddress.address:
				if !ok {
					org.Terminate()
					return
				}
				method, ok := planningMethodsMap[mail.recipientServerName]
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
			case function, _ := <-T_T.updateNotify:
				function()
			}
		}
	}()
}
