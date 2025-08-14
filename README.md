# Project go-redisbenchmark

One Paragraph of project description goes here

## Getting Started

Support Centos/Ubuntu/Debain/Windows/Macos

## MakeFile

Run build make command with tests
```bash
make all
```

Build the application
```bash
make build
```

Cross-platform builds (output to dist/):
```bash
make build-all
```

Run the test suite:
```bash
make test
```

Clean up binary from the last build:
```bash
make clean
```

## Run
- Run via Go directly (cross-platform):
```bash
go run cmd/redis-benchmark/main.go --read-all -n 100000 -c 200 -P 50
```

- Run built binary (Windows):
```powershell
./redis-benchmark.exe --read-all -n 100000 -c 200 -P 50
```

- Run built binary (Linux/macOS):
```bash
./redis-benchmark --read-all -n 100000 -c 200 -P 50
```

## Live reload (optional)
With air installed or via Makefile helper:
```bash
make watch
```
This will run `go run cmd/redis-benchmark/main.go` on changes.
