package servicecenterv3_test

import (
	"fmt"

	servicecenterv3 "gateway/internal/servicecenterv3"
	"gateway/internal/servicecenterv3/contract"
)

func ExampleEngineV3() {
	fmt.Println(servicecenterv3.EngineV3)
	// Output: servicecenterv3
}

func ExampleCallContext_RequireCenter() {
	cc := contract.CallContext{
		TenantID:           "acme",
		CenterInstanceName: "hub-1",
	}
	fmt.Println(cc.RequireCenter() == nil)
	// Output: true
}
