package goplug

type OnOneShot func(args []string) error

// Host has to be implemented and passed to GoPlug by the host application.
type Host interface {
	// RegisterOneShot will be called if a plugin provides OneShot functionality.
	// The action should be used to "start" a OneShot plugin.
	// For example the implementation can register a
	// subcommand based on the Metadata field of the info.
	// When this subcommand gets called, the action has to be executed.
	RegisterOneShot(info PluginInfo, action OnOneShot) error
}
