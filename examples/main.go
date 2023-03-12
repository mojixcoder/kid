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

func helloHandler(c *kid.Context) error {
	return c.JSON(http.StatusOK, kid.Map{"message": "Hello Kid!"})
}
