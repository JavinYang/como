package como

import (
	//	"fmt"
	"reflect"
	"time"
)

// 所有公约
type pacts struct {
	StaticOrg  *staticOrg
	DynamicOrg *dynamicOrg
}

// 组织规定规范
type provision interface {
	init(pactRegisterName string, mailLen int, overTime *time.Duration) (newMailBoxAddress chan mail)
	deliverMailForMailBox(newMail mail)
	Init()
	Info()
	RoutineStart()
	RoutineEnd()
	Terminate()
}

// 静态组织公约
type staticOrg struct {
	OrgsMailBoxAddress map[string]chan mail
}

// 加入静态组织
func (this *staticOrg) Join(registerName string, provision provision, mailLen int) {
	_, ok := this.OrgsMailBoxAddress[registerName]
	if ok {
		panic("已经存在叫做" + registerName + "的静态组织")
		return
	}

	newMailBoxAddress := provision.init(registerName, mailLen, nil)
	this.OrgsMailBoxAddress[registerName] = newMailBoxAddress

	provision.Init()

	rProvision := reflect.ValueOf(provision)

	planningMethodsMap := make(map[string]func())
	numMethod := rProvision.NumMethod()
	for i := 0; i < numMethod; i++ {
		methodName := rProvision.Type().Method(i).Name
		switch methodName {
		case "init":
		case "Init":
		case "Info":
		case "Routine":
		case "Terminate":
		default:
			planningMethodsMap[methodName] = rProvision.Method(i).Interface().(func())
		}
	}

	go func() {
		for {
			select {
			case v, _ := <-newMailBoxAddress:
				method, ok := planningMethodsMap[v.SenderName]
				if !ok {
					v.acceptLine <- false
					continue
				}
				provision.deliverMailForMailBox(v)
				provision.RoutineStart()
				// 如果被领导回绝就continue 不能向下走了
				v.acceptLine <- true
				method()
				provision.RoutineEnd()
			}
		}
	}()
}

// 查询静态组织邮箱地址
func (this *staticOrg) FindMailBoxAddress(RegisterName string) chan mail {
	return nil
}

// 动态组织公约
type dynamicOrg struct {
	Orgs map[string]provision
}

// 加入动态组织
func (this *dynamicOrg) Join(RegisterName string, provision provision, overtime time.Duration) {}

// 获取新动态组织
func (this *dynamicOrg) New(draft draft) chan mail {
	return nil
}
