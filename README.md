# envsubst

[![Go Reference](https://pkg.go.dev/badge/github.com/ninlil/envsubst.svg)](https://pkg.go.dev/github.com/ninlil/envsubst)
[![Go Report Card](https://goreportcard.com/badge/github.com/ninlil/envsubst)](https://goreportcard.com/report/github.com/ninlil/envsubst)
[![License](https://img.shields.io/github/license/ninlil/envsubst)](LICENSE)

A high-performance, stream-oriented, and highly customizable environment variable substitution template engine for Go.

Unlike the standard library's `os.Expand` or simpler string-replacers, `envsubst` is built for low-memory overhead, streaming raw data (`io.Reader` and `io.Writer`), and safety in concurrent environments.

## ✨ Features

* **Low Memory Footprint:** Processes data via buffered streams (using `bufio`). Perfect for large configuration files or massive templates without swallowing memory.
* **Flexible Input Formats:** Seamlessly handles `string`, `[]byte`, or stream-oriented `io.Reader` inputs.
* **Thread-Safe by Design:** Use `Replacer` instances for isolated, concurrent-safe settings context across goroutines.
* **Custom Syntax/Delimiters:** Choose from multiple prefix characters (`$`, `%`, `#`, `&`) & wrapper types (`()`, `{}`, `[]`, `<>`).
* **Robust Variable Lookup Control:** Formulate strict templating errors using `LookupEnv` (fails on missing variables) or non-strict fallback with standard `Getenv`.
* **Zero Dependencies:** Pure Go standard library under the hood.

## 🚀 Installation

```sh
go get github.com/ninlil/envsubst
```

## 📖 Quick Start

The quickest way to get started is using the package-level helpers.

### 1. Basic String Substitution (Standard Lookup)

By default, missing parameters are silently replaced with an empty string, mimicking standard shells.

```go
package main

import (
  "fmt"
  "github.com/ninlil/envsubst"
)

func main() {
  // Simple lookup defaults to envsubst.Getenv when nil is passed
  out, err := envsubst.ConvertString("Hello $(USER)!", nil)
  if err != nil {
    panic(err)
  }
  fmt.Println(out)
}
```

### 2. Strict Checking (Fails on Missing Variables)

If template validation is essential, use `envsubst.LookupEnv`. It returns an error if any referenced variable inside the template is missing in the environment.

```go
package main

import (
  "fmt"
  "github.com/ninlil/envsubst"
)

func main() {
  _, err := envsubst.ConvertString("Database host is $(DB_HOST)", envsubst.LookupEnv)
  if err != nil {
    fmt.Println("Error:", err) // Error: field 'DB_HOST' is missing
  }
}
```

### 3. Custom Map Substitution

You can substitute using static values or key-value structures using `envsubst.Map`.

```go
package main

import (
  "fmt"
  "github.com/ninlil/envsubst"
)

func main() {
  data := map[string]string{
    "PROJECT": "envsubst",
    "VERSION": "v1.2.0",
  }

  out, err := envsubst.ConvertString("Building $(PROJECT) ($(VERSION))", envsubst.Map(data))
  if err == nil {
    fmt.Println(out) // Building envsubst (v1.2.0)
  }
}
```

---

## 🛠️ Advanced Usage

### Concurrency-Safe isolated `Replacer` (Recommended)

When working across multiple goroutines or requiring custom delimiters, always use `NewReplacer` instead of modifying package-level defaults. `Replacer` holds configuration in an immutable struct context after instantiation, allowing multiple goroutines to call `Convert`, `ConvertBytes`, or `ConvertString` safely and concurrently.

```go
package main

import (
  "fmt"
  "github.com/ninlil/envsubst"
)

func main() {
  // Configure with a custom prefix and brackets
  replacer := envsubst.NewReplacer(
    envsubst.WithPrefix('%'),
    envsubst.WithWrapper('{'), // Supports standard bracket groups
  )

  out, err := replacer.ConvertString("Running custom configuration %{CONFIG_NAME}", envsubst.Getenv)
  if err != nil {
    panic(err)
  }
  fmt.Println(out)
}
```

### Buffered Stream Processing (`io.Reader` and `io.Writer`)

Do not load massive files into memory. With `Convert`, process template configurations directly from file handles or socket streams into buffer structures:

```go
package main

import (
  "bytes"
  "fmt"
  "os"
  "github.com/ninlil/envsubst"
)

func main() {
  inputFile, err := os.Open("config.template.yaml")
  if err != nil {
    panic(err)
  }
  defer inputFile.Close()

  var output bytes.Buffer

  // Processes template efficiently with low-alloc stream conversion
  err = envsubst.Convert(inputFile, &output, envsubst.LookupEnv)
  if err != nil {
    panic(err)
  }

  fmt.Println(output.String())
}
```

---

## ⚙️ Configuration Matrix

### Supported Prefix Delimiters

Valid characters to signify a variable start (default represents `'$'`):

* `$`
* `%`
* `#`
* `&`

### Supported Wrapper Types

Specify wrappers to capture variable names. You can supply either the opening or closing character of the pair (default is `()`):

* `()`
* `{}`
* `[]`
* `<>`

---

## 🆚 Comparison with `os.Expand`

| Feature | `os.Expand` / `os.ExpandEnv` | `github.com/ninlil/envsubst` |
| --- | --- | --- |
| **Data Types** | `string` | `string`, `[]byte`, and zero-alloc streams (`io.Reader` / `io.Writer`) |
| **Buffering** | ❌ None | `bufio` based low-allocation buffers |
| **Strict Checking** | ❌ No, forces empty string replacement | Yes, using `LookupEnv` or errors on missing values |
| **Custom Prefix** | ❌ No, fixed to `$` | Yes (`$`, `%`, `#`, `&`) |
| **Custom Wrapper** | ❌ No, fixed to `{}` or none | Yes (`()`, `{}`, `[]`, `<>`) |
| **Thread Safety** | ❌ Global functions only | Yes, using `NewReplacer` instances |

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
