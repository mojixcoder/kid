package main

import (
	"net/http"

	"github.com/mojixcoder/kid"
)

func main() {
	k := kid.New()

	k.Get("/greet", greetHandler)

	k.Run()
}

func greetHandler(c *kid.Context) {
	c.String(http.StatusOK, "Hello World!")
}
