package comm

import (
	"time"
)

// 所有条约
type Pacts struct {
	StaticOrg  StaticOrg
	DynamicOrg DynamicOrg
}

// 条约范本
type Pact struct{ Orgs map[string]OrgPlanning }

// 动态组织条约
type StaticOrg Pact

type Planning interface {
	init()
}

// 加入静态组织
func (this *StaticOrg) Join(RegisterName string, planning Planning) {
	planning.init()
}

// 查询静态组织邮箱地址
func (this *StaticOrg) FindMailBoxAddress(RegisterName string) chan Mail {
	return nil
}

// 静态组织条约
type DynamicOrg Pact

// 加入动态组织
func (this *DynamicOrg) Join(RegisterName string, planning Planning, overtime time.Duration) {}

// 获取新动态组织
func (this *DynamicOrg) New(draft Draft) chan Mail {
	return nil
}
