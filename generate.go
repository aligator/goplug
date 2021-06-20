//go:generate go run ./generate -o ./example/host/gen ./example/host
//go:generate go build -o ./example/plugin-bin  ./example/plugin
//go:generate go build -o ./example/plugin-bin  ./example/plugin2
package goplug
