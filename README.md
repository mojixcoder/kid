![test](https://github.com/mojixcoder/kid/actions/workflows/test.yml/badge.svg)
[![code quality](https://app.codacy.com/project/badge/Grade/aa9e650027e144359ae6f3cbdcdae6c9)](https://www.codacy.com/gh/mojixcoder/kid/dashboard?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=mojixcoder/kid&amp;utm_campaign=Badge_Grade)
[![coverage](https://app.codacy.com/project/badge/Coverage/aa9e650027e144359ae6f3cbdcdae6c9)](https://www.codacy.com/gh/mojixcoder/kid/dashboard?utm_source=github.com&utm_medium=referral&utm_content=mojixcoder/kid&utm_campaign=Badge_Coverage)

### Kid Micro Web Framework
___
**Kid** is a micro web framework written in Go. It aims to keep its core simple and yet powerful.

#### Quick Start
___

To install Kid Go 1.18 or higher is required: `go get github.com/mojixcoder/kid`

Create `server.go`:

```go
package main

import (
    "net/http"

    "github.com/mojixcoder/kid"
)

func main() {
    k := kid.New()

    k.GET("/hello", helloHandler)

    k.Run()
}

func helloHandler(c *kid.Context) error {
    return c.JSON(http.StatusOK, kid.Map{"message": "Hello Kid!"})
}
```

#### TODOs
___

- [x] Add test cases up to +90% coverage.
- [ ] Complete docs.
- [x] Add more methods for sending response like XML, HTML, etc.
- [ ] Add some middlewares like `Logger`, `Recovery`, etc.
- [x] Add CI.
- [x] Add comments.
- [x] Add methods to serve static files.
- [ ] Add validator.
