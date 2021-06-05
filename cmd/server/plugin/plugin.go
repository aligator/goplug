package plugin

import "goplug"

// TestPlugin defines the methods which can be used by plugins.
type TestPlugin struct {
	goplug.PluginImpl
}

func (p TestPlugin) Print(message string) {
	p.Send("print", message)
}
