package main

import (
	"viamroomba"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
	base "go.viam.com/rdk/components/base"
)

func main() {
	// ModularMain can take multiple APIModel arguments, if your module implements multiple models.
	module.ModularMain(resource.APIModel{ base.API, viamroomba.Base})
}
