///-------------------------------------------------------------------------------
///                               COMM 0.1
///-------------------------------------------------------------------------------

package comm

import "fmt"

var PactOrg *pacts

func init() {
	fmt.Println("11")
	createPactOrg()
}

func createPactOrg() {
	staticOrg := &staticOrg{OrgsMailBoxAddress: make(map[string]chan mail)}
	dynamicOrg := &dynamicOrg{Orgs: make(map[string]planning)}
	PactOrg = &pacts{staticOrg, dynamicOrg}
}

///-------------------------------------------------------------------------------
///  Javin Yang <120696788@qq.com>
///  2017-5-06
///  2017-5-07
///-------------------------------------------------------------------------------
