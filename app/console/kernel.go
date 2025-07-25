package console

import (
	"flag"
	"src/app/console/commands"
	"src/app/console/getters"

	"github.com/confetti-framework/framework/foundation/console"
	"github.com/confetti-framework/framework/inter"
)

// NewKernel ensures that the kernel receives all existing commands
// and that the correct flag.Getters are used.
func NewKernel(app inter.App) console.Kernel {
	return console.Kernel{
		App: app,

		// Here you can add your own commands.
		Commands: []inter.Command{
			commands.Watch{},
			commands.PkgPull{},
			commands.PkgPush{},
			commands.ContainerQuery{},
		},

		// This list includes custom flag.Getters, you can create custom
		// types to cast flags from the command to a value.
		Getters: []flag.Getter{
			new(getters.StringList),
			new(getters.IntList),
		},
	}
}
