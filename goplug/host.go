package goplug

type OnOneShot func(args []string) error

type Host interface {
	RegisterOneShot(info PluginInfo, action OnOneShot) error
}
