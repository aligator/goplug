# GoPlug

A simple Go plugin mechanism which just relies on starting plugins as
child processes and communicates with them over stdin / stdout.

## Run example
First build the plugin
```bash
go build -o ./cmd/plugin-bin ./cmd/plugin
```

Then run the 'server'
```bash
go run ./cmd/server
```