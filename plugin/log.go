package plugin

// LogWriter can be used to configure any logger to write
// 'log' commands, as stdout cannot be used by plugins.
type LogWriter struct {
	Plugin *Plugin
}

func (l LogWriter) Write(p []byte) (int, error) {
	err := l.Plugin.Send("log", string(p))
	if err != nil {
		return 0, err
	}

	return len(p), nil
}
