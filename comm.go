///-------------------------------------------------------------------------------
///                              COMM 0.1
///-------------------------------------------------------------------------------

package comm

import "fmt"

var PactOrg *pacts

func init() {
	fmt.Println("11")
	createPactOrg()
}

func createPactOrg() {
	staticOrg := &staticOrg{orgs: make(map[string]OrgPlanning)}
	dynamicOrg := &dynamicOrg{orgs: make(map[string]OrgPlanning)}
	PactOrg = &pacts{staticOrg, dynamicOrg}
}

///-------------------------------------------------------------------------------
///  Javin Yang <120696788@qq.com>
///  2017-5-06
///-------------------------------------------------------------------------------
