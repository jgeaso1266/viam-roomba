package main

import (
	"io"
	"log"
	viamroomba "viamroomba"

	base "go.viam.com/rdk/components/base"
	"go.viam.com/rdk/components/sensor"
	"go.viam.com/rdk/module"
	"go.viam.com/rdk/resource"
)

func main() {
	// The go-roomba library calls log.Printf on every serial write, which Go
	// sends to stderr. Viam treats all stderr output from modules as errors.
	// Discard the default logger to suppress that noise.
	log.SetOutput(io.Discard)

	module.ModularMain(
		resource.APIModel{API: base.API, Model: viamroomba.Base},
		resource.APIModel{API: sensor.API, Model: viamroomba.Sensor},
	)
}
