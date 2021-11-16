package provision

import (
	"github.com/docker/machine/libmachine/auth"
	"github.com/docker/machine/libmachine/engine"
)
// 	"github.com/docker/machine/libmachine/bootstrap"
type EngineConfigContext struct {
	DockerPort       int
	AuthOptions      auth.Options
	EngineOptions    engine.Options
	DockerOptionsDir string
}
//	BoostrapOptions    bootstrap.Options