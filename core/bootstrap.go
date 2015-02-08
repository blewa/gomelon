// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package core

// Bootstrap contains everything required to bootstrap a command
type Bootstrap struct {
	Application Application
	Arguments   []string

	ConfigurationFactory ConfigurationFactory
	ServerFactory        ServerFactory

	bundles  []Bundle
	commands []Command
}

// NewBootstrap allocates and returns a new Bootstrap.
func NewBootstrap(app Application) *Bootstrap {
	bootstrap := &Bootstrap{
		Application: app,
	}
	return bootstrap
}

// Bundles returns registed bundles.
func (bootstrap *Bootstrap) Bundles() []Bundle {
	return bootstrap.bundles
}

// AddBundle adds the given bundle to the bootstrap. AddBundle is not concurrent-safe.
func (bootstrap *Bootstrap) AddBundle(bundle Bundle) {
	bundle.Initialize(bootstrap)
	bootstrap.bundles = append(bootstrap.bundles, bundle)
}

// Commands returns registed commands.
func (bootstrap *Bootstrap) Commands() []Command {
	return bootstrap.commands
}

// AddCommand add the given command to the bootstrao. AddCommand is not concurrent-safe.
func (bootstrap *Bootstrap) AddCommand(command Command) {
	bootstrap.commands = append(bootstrap.commands, command)
}

// run runs all registered bundles
func (bootstrap *Bootstrap) Run(configuration *Configuration, environment *Environment) error {
	for _, bundle := range bootstrap.bundles {
		if err := bundle.Run(configuration, environment); err != nil {
			return err
		}
	}
	return nil
}