package goplug

// PluginLogWriter can be used to configure any logger to write
// 'log' commands, as stdout cannot be used by plugins.
type PluginLogWriter struct {
	Plugin *Plugin
}

func (l PluginLogWriter) Write(p []byte) (int, error) {
	err := l.Plugin.Send("log", string(p))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
