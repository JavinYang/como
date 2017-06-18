package como

import (
	"reflect"
	"time"
)

// 公约实例
var Pact *pacts

// 初始化公约实例
func init() {
	staticOrg := &staticPact{orgsMailBoxAddress: make(map[string]MailBoxAddress)}
	dynamicOrg := &dynamicPact{orgsProvision: make(map[string]dynamicOrgInfo)}
	Pact = &pacts{staticOrg, dynamicOrg}
}

// 所有公约
type pacts struct {
	Static  *staticPact
	Dynamic *dynamicPact
}

// 组织规定规范
type provision interface {
	init(pactRegisterName string, mailLen int, overTime *int64) (newMailBoxAddress MailBoxAddress)
	deliverMailForMailBox(newMail mail)
	getLeader() *leader
	Init(...interface{})
	Info()
	routineStart()
	RoutineStart()
	RoutineEnd()
	Terminate()
}

// 静态组织公约
type staticPact struct {
	orgsMailBoxAddress map[string]MailBoxAddress
}

// 公开加入静态组织
func (this *staticPact) Join(registerName string, org provision, mailLen int, initPars ...interface{}) {
	_, ok := this.orgsMailBoxAddress[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的静态组织")
		return
	}

	newMailBoxAddress := org.init(registerName, mailLen, nil)
	this.orgsMailBoxAddress[registerName] = newMailBoxAddress

	this.run(newMailBoxAddress, org, initPars...)
}

// 匿名加入静态组织
func (this *staticPact) JoinAnonymous(org provision, mailLen int, initPars ...interface{}) {
	newMailBoxAddress := org.init("", mailLen, nil)
	this.run(newMailBoxAddress, org, initPars...)
}

// 运行静态组织
func (this *staticPact) run(newMailBoxAddress MailBoxAddress, org provision, initPars ...interface{}) {
	orgReflect := reflect.ValueOf(org)

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
					T_T.goodByeMyFriends()
					return
				}
				method, ok := planningMethodsMap[mail.recipientName]
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

// 查询静态组织邮箱地址
func (this *staticPact) FindMailBoxAddress(RegisterName string) (mailBoxAddress MailBoxAddress, ok bool) {
	mailBoxAddress, ok = this.orgsMailBoxAddress[RegisterName]
	return
}

// 动态组织公约
type dynamicPact struct {
	orgsProvision map[string]dynamicOrgInfo
}

// 动态组织信息
type dynamicOrgInfo struct {
	orgType  reflect.Type
	mailLen  int
	overtime int64
}

// 加入动态组织
func (this *dynamicPact) Join(registerName string, provision provision, mailLen int, overtime int64) {

	if overtime < 0 || overtime == 0 {
		panic(registerName + "加入动态组织的overtime不能<=0因为着没有意义")
	}

	_, ok := this.orgsProvision[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的动态组织")
	}
	this.orgsProvision[registerName] = dynamicOrgInfo{reflect.Indirect(reflect.ValueOf(provision)).Type(), mailLen, overtime}
}

// 用名字生成动态组织
func (this *dynamicPact) Generate(registerName string, initPars ...interface{}) (mailBoxAddress MailBoxAddress, ok bool) {
	dynamicOrgInfo, ok := this.orgsProvision[registerName]
	if !ok {
		return MailBoxAddress{}, false
	}
	overtime := dynamicOrgInfo.overtime
	orgReflect := reflect.New(dynamicOrgInfo.orgType)
	org := orgReflect.Interface().(provision)
	newMailBoxAddress := org.init(registerName, dynamicOrgInfo.mailLen, &overtime)
	this.run(newMailBoxAddress, orgReflect, org, overtime, initPars...)
	return newMailBoxAddress, true
}

// 创建一个新动态组织
func (this *dynamicPact) New(org provision, mailLen int, overtime int64, initPars ...interface{}) (mailBoxAddress MailBoxAddress, ok bool) {
	newMailBoxAddress := org.init("", mailLen, &overtime)
	orgReflect := reflect.ValueOf(org)
	this.run(newMailBoxAddress, orgReflect, org, overtime, initPars...)
	return newMailBoxAddress, true
}

// 运行动态组织
func (this *dynamicPact) run(newMailBoxAddress MailBoxAddress, orgReflect reflect.Value, org provision, overtime int64, initPars ...interface{}) {
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
					T_T.goodByeMyFriends()
					return
				}

				method, ok := planningMethodsMap[mail.recipientName]
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
