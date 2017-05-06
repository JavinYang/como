package comm

import (
	"time"
)

// 所有条约
type pacts struct {
	StaticOrg  *staticOrg
	DynamicOrg *dynamicOrg
}

// 条约范本
type pact struct{ orgs map[string]OrgPlanning }

// 动态组织条约
type staticOrg pact

type planning interface {
	init(*string, *time.Duration, *map[chan interface{}]interface{}, *leader, *mailBox)
	Start()
	Info()
	Routine()
	Terminate()
}

// 加入静态组织
func (this *staticOrg) Join(RegisterName string, planning planning) {
	planning.Start()
}

// 查询静态组织邮箱地址
func (this *staticOrg) FindMailBoxAddress(RegisterName string) chan mail {
	return nil
}

// 静态组织条约
type dynamicOrg pact

// 加入动态组织
func (this *dynamicOrg) Join(RegisterName string, planning planning, overtime time.Duration) {}

// 获取新动态组织
func (this *dynamicOrg) New(draft draft) chan mail {
	return nil
}
