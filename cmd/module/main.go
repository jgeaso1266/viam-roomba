package main

import (
	viamroomba "viamroomba"

	base "go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	module.ModularMain(
		resource.APIModel{API: base.API, Model: viamroomba.Base},
		resource.APIModel{API: sensor.API, Model: viamroomba.Sensor},
	)
}
