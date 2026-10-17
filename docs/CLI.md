# Decoy CLI

A CLI tool for generating and ingesting mock data using Go templates and configurable runners.


## Features

* Template Engine - Generate dynamic data using Go's `text/template`. See [Template Engine](#template-engine) section.
    * Use built-in functions like random generation, probability, counters, and more.
* Runner Engine - Ingest generated data using implemented Runners. See [Runner Engine](#runner-engine) section.
    * Concurrent Execution - You can execute the runner `n` times using multiple goroutines.
* Persistence - You can save Templates and Runners to reuse later.
* Mock server - Create mock servers using the Template Engine and the OpenAPI Specification.


## Installation

### Install from source

```bash
go install github.com/aaron70/decoy-cli/decoy@latest
```

### Build locally

```bash
git clone https://github.com/aaron70/decoy-cli.git
cd decoy-cli
go build -o decoy ./decoy/main.go
```


## Template Engine

Templates use Go's [`text/template`](https://pkg.go.dev/text/template) syntax with additional functions from the [decoy](https://github.com/aaron70/decoy) library.

### Built-in template functions

> **Sprig functions:** All [Sprig v3](http://masterminds.github.io/sprig/) template functions (date formatting, string manipulation, math, crypto, type conversion, etc.) are also available alongside the functions below. For example, Sprig provides `list`, `coalesce`, `env`, `fromJson`, and many other utilities.

#### Random generation

| Function | Description | Example |
|---|---|---|
| `randomInt min max` | Random int in `[min, max)` | `{{randomInt 1 100}}` |
| `randomFloat min max` | Random float64 in `[min, max)` | `{{randomFloat 0.0 1.0}}` |
| `randomBoolean` | Random true/false | `{{randomBoolean}}` |
| `randomChoice args...` | Random item from args | `{{randomChoice "a" "b" "c"}}` |
| `randomChoiceList choices` | Random item from a slice | `{{randomChoiceList $colors}}` |
| `randomText maxWords` | Markov chain random text | `{{randomText 50}}` |
| `randomName` | Random first name | `{{randomName}}` |
| `randomLastName` | Random last name | `{{randomLastName}}` |
| `randomFullName prob` | Random full name with middle name probability | `{{randomFullName 0.5}}` |

#### Probability

| Function | Description | Example |
|---|---|---|
| `probability p` | Returns true with probability `p` | `{{probability 0.75}}` |

#### Incremental counters

| Function | Description | Example |
|---|---|---|
| `nextIncrementalInt id start step` | Next value of named counter | `{{nextIncrementalInt "c" 1 1}}` |
| `currentIncrementalInt id default` | Current value (non-advancing) | `{{currentIncrementalInt "c" 0}}` |
| `setIncrementalInt id value` | Set a named counter to a value | `{{setIncrementalInt "c" 42}}` |
| `unsetIncrementalInt id` | Remove a named counter | `{{unsetIncrementalInt "c"}}` |

#### Pagination

| Function | Description | Example |
|---|---|---|
| `paginationPage page size total` | Paginate by page number; returns `{ids, page, size, total}` | `{{paginationPage 1 20 100}}` |
| `paginationSkip skip limit total` | Paginate by offset; returns `{ids, skip, limit, total}` | `{{paginationSkip 20 20 100}}` |
| `paginationToken token limit total` | Paginate by next-token (base64); returns `{ids, nextToken, limit, total}` | `{{paginationToken "" 20 100}}` |

#### I/O

| Function | Description | Example |
|---|---|---|
| `readFileString path` | Read file as string | `{{readFileString "data.txt"}}` |
| `readFileBytes path` | Read file as bytes | `{{readFileBytes "img.png"}}` |
| `readFileBase64 path` | Read file as base64 | `{{readFileBase64 "img.png"}}` |

See [decoy's FUNCTIONS.md](FUNCTIONS.md) for full details.


## Runner Engine

Decoy provides some Runners to allow you to parse and ingest the data generated to your applications.

You can execute a Runner `n` times with `g` concurrent Goroutines, to generate multiple records with the given runner and given template.
The runner's configuration support the [Template Engine](#template-engine) so you can use the built-in functions within the configuration.
Additionally, Runners provide the following data available to the Template Engine:

| Key | Description | Example |
| --- | ----------- | ------- |
| times | The number of times that the runner will be executed | `{{ .times }}` |
| goroutines | The number of concurrent goroutines that will execute the runner | `{{ .goroutines }}` |
| template | The contents of the already parsed template | `{{ .template }}` |
| data | The JSON data passed to the template | `{{ .data }}` |


### Runner types

#### `http` — HTTP requests

Make HTTP requests with configurable method, URL, query parameters, headers, and body.

Configuration:

The configuration for the `http` runner is a JSON object with the HTTP request configuration.

```json
{
  "method": "<HTTP Method>",
  "url": "<URL>",
  "queryParameters": <JSON Object>,
  "headers": <JSON Object>,
  "body": <Any JSON Node>
}
```

Configuration example:
```json
{
  "method": "POST",
  "url": "http://api.example.com/items",
  "queryParameters": {"source": "decoy"},
  "headers": {"Authorization": "Bearer token123"},
  "body": {{.template}}
}
```

#### `cmd` — Shell commands

Executes the shell command and prints the results from stdout and stderr. 

Configuration:

The configuration for the `cmd` runner is the command itself to run. 

Configuration example:
```text
echo "{{.template}}"
```


## Advanced template example

This is a demonstration of how you can use the provided functions to build a fairly complex example.

`~/templates/user.tmpl`
```text
{{- $countries := list "Costa Rica" "United States" "Panama" -}}
{{- $isAdult := probability 0.70 -}}
{{- $isGranny := and $isAdult (probability 0.20) -}}
{{- $country := randomChoiceList $countries | coalesce .country -}}
{{- $siblingsCount := 0 -}}
{{- if probability 0.675 -}} {{/* Has siblings or not */}}
  {{- if probability 0.35 -}} {{/* Has more than 5 */}}
    {{- $siblingsCount = randomInt 5 16 -}}
  {{- else -}}
    {{- $siblingsCount = randomInt 1 5 -}}
  {{- end -}}
{{- end -}}
{{- $isMarried := and $isAdult (probability 0.43) -}}

{
  "id": {{ nextIncrementalInt "id" 0 1 }},
  "name": "{{ randomFullName 0.85 }}",
  "age": {{ if not $isAdult -}}
    {{- randomInt 0 18 -}}
  {{- else -}}
    {{- if $isGranny -}}
      {{- randomInt 60 120 -}}
    {{- else -}}
      {{- randomInt 18 60 -}}
    {{- end -}}
  {{- end -}},
  "adult": {{ $isAdult }},
  "favoriteQuote": "{{ randomText 15 }}...",
  "country": "{{ $country }}",
  "sex": "{{ randomChoice "Female" "Male" }}",
  "married": {{ $isMarried }},
  {{- if $isMarried }}
  "spouse": "{{ randomFullName 0.85 }}",
  {{- end }}
  "siblings": [
  {{- range $i := $siblingsCount }}
    "{{ randomFullName 0.85 }}"
    {{- if not (eq $i (sub $siblingsCount 1)) -}},{{ end -}}
  {{ end }}
  ]
}
```

```bash
decoy run cmd -c "echo {{ .template }}" -f ~/templates/user.tmpl -v "country=Costa Rica" -n 1
```

The resulting user would be similar to:

```json
{
  "id": 0,
  "name": "Margaret Pham Ross",
  "age": 61,
  "adult": true,
  "favoriteQuote": "me simply a motive in art. You might see nothing in him. I see everything...",
  "country": "Costa Rica",
  "sex": "Female",
  "married": true,
  "spouse": "Priya Reyes Mensah",
  "siblings": [
    "Olga Jung Bell"
  ]
}
```
