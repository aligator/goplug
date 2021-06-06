# GoPlug

A simple (currently POC only, so not for production use) Go plugin mechanism which just relies on starting plugins as
child processes and communicates with them over stdin / stdout.

## Run example
First build the plugin
```bash
go build -o ./cmd/plugin-bin ./cmd/plugin  ./cmd/plugin2
```

Then run the 'server'
```bash
go run ./cmd/server
```