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

// 组织计划规范
type planning interface {
	init(pactRegisterName string, overTime *time.Duration)
	Init()
	Info()
	Routine()
	Terminate()
}

// 静态组织条约
type staticOrg pact

// 加入静态组织
func (this *staticOrg) Join(RegisterName string, planning planning) {
	planning.Init()
}

// 查询静态组织邮箱地址
func (this *staticOrg) FindMailBoxAddress(RegisterName string) chan mail {
	return nil
}

// 动态组织条约
type dynamicOrg pact

// 加入动态组织
func (this *dynamicOrg) Join(RegisterName string, planning planning, overtime time.Duration) {}

// 获取新动态组织
func (this *dynamicOrg) New(draft draft) chan mail {
	return nil
}
