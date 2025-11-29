<div align="center">
    <h2>QCL</h2>
    <h5>a simple language that allows you to check the eval result of a query. </h5>
</div>

## Intro

It's designed to be used in ACL (Access Control List) systems, where you need to check if a user has access to a resource.

### Example

```js
(@record.published || @record.owner == @req.user.id) // Normal case
|| // Or operator
(@req.user.role == 'admin' || @req.user.id in @record.granted) // Special case
```

Let's break it down:

- `@req.user.role == 'admin'`: Check if the user has the role of `admin`.
- `@req.user.id in @record.granted`: Check if the user's id is in the `granted` list of the record.
- `@record.published`: Check if the record is published.
- `@record.owner == @req.user.id`: Check if the record's owner is the user.

The above example is a simple ACL system that checks if the user has access to a record.

More language details can be found in [LANG.md](LANG.md).

## Features

- `json` (enabled by default)
- `yaml`
- `toml`

At least one input format feature must be enabled. The default configuration enables `json`.

### Usage

#### Integration (Go)

```go
package main

import (
    "fmt"

    "gqcl"
)

func main() {
    expr, err := gqcl.Parse("@req.user.role == \"admin\" || @record.published")
    if err != nil {
        panic(err)
    }

    ctx := gqcl.FromInterface(map[string]any{
        "req": map[string]any{
            "user": map[string]any{
                "role": "admin",
                "id":   "u1",
            },
        },
        "record": map[string]any{
            "owner":     "u2",
            "published": false,
            "granted":   []any{"u3", "u1"},
        },
    })

    val, err := expr.Eval(ctx)
    if err != nil {
        panic(err)
    }

    fmt.Println(val.String()) // true
}
```

#### CLI

```bash
cat context.json | go run ./cmd/qcl "@req.user.id == 'alice' || @record.published"
cat context.yaml | go run ./cmd/qcl --yaml "@roles.0 == 'admin'"
```

<div height="100px" align="center">
    <img src="https://cdn.lpkt.cn/img/capture/qcl.png" alt="QCL" />
</div>

## Ports

- [Rust SDK / CLI](https://github.com/lollipopkit/qcl)

## License

```plaintext
Apache-2.0 lollipopkit
```
