# Decoy

A Go library and CLI tool for generating and ingesting mock/synthetic data using Go templates.

## CLI

### Features

* Template Engine - Generate dynamic data using Go's `text/template`. See [Template Engine](docs/CLI.md#template-engine) section.
    * Use built-in functions like random generation, probability, counters, and more.
* Runner Engine - Ingest generated data using implemented Runners. See [Runner Engine](docs/CLI.md#runner-engine) section.
    * Concurrent Execution - You can execute the runner `n` times using multiple goroutines.
* Persistence - You can save Templates and Runners to reuse later.
* Mock server - Create mock servers using the Template Engine and the OpenAPI Specification.


### Installation

```bash
go install github.com/aaron70/decoy/cmd/decoy@latest
```

Or build from source:

```bash
git clone https://github.com/aaron70/decoy.git
cd decoy
go build -o decoy ./cmd/decoy/main.go
```


### Quick Start

#### Store a template

```bash
decoy template store greet -t 'Hello, {{ coalesce .Name "World" }}!'
```

#### Parse the template

```bash
decoy template parse greet --data '{ "Name": "Doe" }'
```

#### Store a runner

```bash
decoy runner store echo -c 'echo User said: "{{ .template }}"' 
```

#### Execute the runner

```bash
decoy runner run cmd echo greet -v Name=Doe
```

#### Start a REST server

```bash
decoy server rest start -f ./spec.yaml
```

#### Explore the commands

```bash
decoy template --help
decoy template parse --help
decoy runner --help
decoy runner run --help
decoy server --help
decoy server start --help
```


## Library Usage

```go
package main

import (
    "fmt"
    "github.com/aaron70/decoy/pkg/decoy"
)

func main() {
    d, _ := decoy.NewDecoyWithSeed(42)

    result, _ := d.ParseTemplateString(
        `{"id": {{nextIncrementalInt "counter" 1 1}}, "name": "{{randomName}} {{randomLastName}}"}`,
    )
    fmt.Println(result)
}
```

All random generation functions (`RandomInt`, `RandomFloat`, `RandomBoolean`, `RandomChoice`, `RandomText`, `RandomName`, etc.) are available as methods on `Decoy` or via the `decoy.RandomChoice` generic function.

## Documentation

- [CLI Reference](docs/CLI.md) — Full command usage, runner engine configuration, advanced examples
- [Template Functions](docs/FUNCTIONS.md) — Complete function signatures, parameters, and examples
- [Server API Reference](docs/API.md) — Decoy server API and mock server.

## Configuration

Templates, runners, and server specs are stored as JSON files under `~/.config/decoy/` (or `$XDG_CONFIG_HOME/decoy`).

