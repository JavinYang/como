package como

import (
	//	"fmt"
	"reflect"
	"time"
)

// 公约实例
var Pact *pacts

// 初始化公约实例
func init() {
	staticOrg := &staticPact{OrgsMailBoxAddress: make(map[string]chan mail)}
	dynamicOrg := &dynamicPact{Orgs: make(map[string]provision)}
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
	Init()
	Info()
	routineStart()
	RoutineStart()
	RoutineEnd()
	Terminate()
}

// 静态组织公约
type staticPact struct {
	OrgsMailBoxAddress map[string]chan mail
}

// 加入静态组织
func (this *staticPact) Join(registerName string, org provision, mailLen int) {
	_, ok := this.OrgsMailBoxAddress[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的静态组织")
		return
	}

	newMailBoxAddress := org.init(registerName, mailLen, nil)
	this.OrgsMailBoxAddress[registerName] = newMailBoxAddress

	org.Init()

	orgReflect := reflect.ValueOf(org)

	planningMethodsMap := make(map[string]func())
	numMethod := orgReflect.NumMethod()
	for i := 0; i < numMethod; i++ {
		methodName := orgReflect.Type().Method(i).Name
		switch methodName {
		case "init":
		case "Init":
		case "Info":
		case "Routine":
		case "Terminate":
		default:
			planningMethodsMap[methodName] = orgReflect.Method(i).Interface().(func())
		}
	}

	go func() {
		for {
			select {
			case v, ok := <-newMailBoxAddress:
				if !ok {
					org.Terminate()
					break
				}
				method, ok := planningMethodsMap[v.SenderName]
				if !ok {
					v.acceptLine <- false
					continue
				}
				org.deliverMailForMailBox(v)
				org.routineStart()
				org.RoutineStart()
				T_T := org.getLeader()
				if !T_T.isAccept() {
					v.acceptLine <- false
					continue
				}
				v.acceptLine <- true
				method()
				org.RoutineEnd()
			}
		}
	}()
}

// 查询静态组织邮箱地址
func (this *staticPact) FindMailBoxAddress(RegisterName string) chan mail {
	return nil
}

// 动态组织公约
type dynamicPact struct {
	Orgs map[string]provision
}

// 加入动态组织
func (this *dynamicPact) Join(RegisterName string, provision provision, overtime time.Duration) {}

// 获取新动态组织
func (this *dynamicPact) New(draft draft) chan mail {
	return nil
}
