
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

func  main() {
    k := kid.New()

    k.GET("/hello", helloHandler)

    k.Run()
}

func  helloHandler(c *kid.Context) error {
    return c.JSON(http.StatusOK, kid.Map{"message": "Hello Kid!"})
}
```

#### TODOs
___

- [ ] Add test cases up to +90% coverage
- [ ] Complete docs
- [ ] Add more methods for sending response like XML, HTML, etc.
- [ ] Add some middlewares like `Logger`, `Recovery`, etc.
- [x] Add CI/CD
- [ ] Add comments
- [ ] Add binder and validator
- [ ] Re-implement router using radix tree
