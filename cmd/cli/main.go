package main

import (
	"context"
	"viamroomba"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
	base "go.viam.com/rdk/components/base"
)

func main() {
	err := realMain()
	if err != nil {
		panic(err)
	}
}

func realMain() error {
	ctx := context.Background()
	logger := logging.NewLogger("cli")

	deps := resource.Dependencies{}
	// can load these from a remote machine if you need

	cfg := viamroomba.Config{}

	thing, err := viamroomba.NewBase(ctx, deps, base.Named("foo"), &cfg, logger)
	if err != nil {
		return err
	}
	defer thing.Close(ctx)

	return nil
}
