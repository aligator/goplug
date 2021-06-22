//go:generate go run ./generate -o gen -p plug -m github.com/aligator/goplug/example/host ./example/host
//go:generate go build -o ./example/plugin-bin  ./example/plugin
//go:generate go build -o ./example/plugin-bin  ./example/plugin2
package main
