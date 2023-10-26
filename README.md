![test](https://github.com/mojixcoder/kid/actions/workflows/test.yml/badge.svg)
[![code quality](https://app.codacy.com/project/badge/Grade/aa9e650027e144359ae6f3cbdcdae6c9)](https://www.codacy.com/gh/mojixcoder/kid/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mojixcoder/kid&amp;utm_campaign=Badge_Grade)
[![coverage](https://app.codacy.com/project/badge/Coverage/aa9e650027e144359ae6f3cbdcdae6c9)](https://www.codacy.com/gh/mojixcoder/kid/dashboard?utm_source=github.com&utm_medium=referral&utm_content=mojixcoder/kid&utm_campaign=Badge_Coverage)

### Kid Micro Web Framework
___
**Kid** is a micro web framework written in Go. It aims to keep its core simple and yet powerful. It's fully compatible with net/http interfaces and can be adapted with other net/http compatible packages as well.

### Features
___
- Robust tree-based router.
- Path parameters.
- Router groups.
- Rich built-in responses(JSON, HTML, XML, string, byte).
- Middlewares.
- Zero dependency, only standard library.
- Compatible with net/http interfaces.
- Extendable, you can also use your own JSON, XML serializers or HTML renderer.

### Versioning
___
This package follows [semver](https://semver.org/) versioning.

#### Quick Start
___

To install Kid Go 1.19 or higher is required: `go get github.com/mojixcoder/kid`

Create `server.go`:

```go
package main

import (
    "net/http"

    "github.com/mojixcoder/kid"
)

func main() {
    k := kid.New()

    k.Get("/hello", helloHandler)

    k.Run()
}

func helloHandler(c *kid.Context) {
    c.JSON(http.StatusOK, kid.Map{"message": "Hello Kid!"})
}
```
