package como

import (
	"reflect"
	"time"
)

// 公约实例
var Pact *pacts

// 初始化公约实例
func init() {
	staticOrg := &staticPact{orgsMailBoxAddress: make(map[string]chan mail)}
	dynamicOrg := &dynamicPact{orgsProvision: make(map[string]provision)}
	Pact = &pacts{staticOrg, dynamicOrg}
}

// 所有公约
type pacts struct {
	Static  *staticPact
	Dynamic *dynamicPact
}

// 组织规定规范
type provision interface {
	init(pactRegisterName string, mailLen int, overTime *time.Duration) (newMailBoxAddress chan mail)
	deliverMailForMailBox(newMail mail)
	getLeader() leader
	Init(...interface{})
	Info()
	routineStart()
	RoutineStart()
	RoutineEnd()
	Terminate()
}

// 静态组织公约
type staticPact struct {
	orgsMailBoxAddress map[string]chan mail
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

	org.Init(initPars)

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
		for {
			select {
			case mail, ok := <-newMailBoxAddress:
				if !ok {
					org.Terminate()
					break
				}
				method, ok := planningMethodsMap[mail.senderName]
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
func (this *staticPact) FindMailBoxAddress(RegisterName string) chan mail {
	mailBoxAddress, ok := this.orgsMailBoxAddress[RegisterName]
	if ok {
		return mailBoxAddress
	}
	return nil
}

// 动态组织公约
type dynamicPact struct {
	orgsProvision map[string]provision
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
	this.orgsProvision[registerName] = provision
	this.mailLen = mailLen
	this.overtime = overtime
}

// 获取新动态组织
func (this *dynamicPact) New(registerName string, initPars ...interface{}) (mailBoxAddress chan mail, ok bool) {
	org, ok := this.orgsProvision[registerName]
	if !ok {
		return nil, false
	}
	overtime := this.overtime
	newMailBoxAddress := org.init(registerName, this.mailLen, &overtime)

	org.Init(initPars)

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
		for {
			select {
			case mail, ok := <-newMailBoxAddress:
				if !ok {
					org.Terminate()
					return
				}
				method, ok := planningMethodsMap[mail.senderName]
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
			case <-time.After(time.Second):
				if overtime == 0 {
					//告诉朋友我要死了
					T_T.Dissolve()
					org.Terminate()
				}
				overtime -= 1
			}
		}
	}()

	return newMailBoxAddress, true
}
