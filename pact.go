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
	dynamicOrg := &dynamicPact{orgsProvision: make(map[string]reflect.Type)}
	Pact = &pacts{staticOrg, dynamicOrg}
}

// 所有公约
type pacts struct {
	Static  *staticPact
	Dynamic *dynamicPact
}

// 组织规定规范
type provision interface {
	init(pactRegisterName string, mailLen int, overTime *time.Duration) (newMailBoxAddress MailBoxAddress)
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

// 加入静态组织
func (this *staticPact) Join(registerName string, org provision, mailLen int, initPars ...interface{}) {
	_, ok := this.orgsMailBoxAddress[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的静态组织")
		return
	}

	newMailBoxAddress := org.init(registerName, mailLen, nil)
	this.orgsMailBoxAddress[registerName] = newMailBoxAddress

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
				method, ok := planningMethodsMap[mail.sendeeName]
				if !ok {
					mail.acceptLine <- false
					continue
				}
				org.deliverMailForMailBox(mail)
				org.routineStart()
				org.RoutineStart()
				if !T_T.isAccept() {
					mail.acceptLine <- false
					continue
				}
				mail.acceptLine <- true
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
	orgsProvision map[string]reflect.Type
	overtime      time.Duration
	mailLen       int
}

// 加入动态组织
func (this *dynamicPact) Join(registerName string, provision provision, mailLen int, overtime time.Duration) {

	if overtime < 0 || overtime == 0 {
		panic(registerName + "加入动态组织的overtime不能<=0因为着没有意义")
	}

	_, ok := this.orgsProvision[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的动态组织")
		return
	}
	this.orgsProvision[registerName] = reflect.Indirect(reflect.ValueOf(provision)).Type()
	this.mailLen = mailLen
	this.overtime = overtime
}

// 获取新动态组织
func (this *dynamicPact) New(registerName string, initPars ...interface{}) (mailBoxAddress MailBoxAddress, ok bool) {
	orgType, ok := this.orgsProvision[registerName]
	if !ok {
		return MailBoxAddress{}, false
	}
	overtime := this.overtime
	orgReflect := reflect.New(orgType)
	org := orgReflect.Interface().(provision)
	newMailBoxAddress := org.init(registerName, this.mailLen, &overtime)

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

				method, ok := planningMethodsMap[mail.sendeeName]
				if !ok {
					mail.acceptLine <- false
					continue
				}
				org.deliverMailForMailBox(mail)
				org.routineStart()
				org.RoutineStart()
				if !T_T.isAccept() {
					mail.acceptLine <- false
					continue
				}
				mail.acceptLine <- true
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
	return newMailBoxAddress, true
}
