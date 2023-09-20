package main

import (
	"fmt"
	"net/http"

	"github.com/mojixcoder/kid"
)

func main() {
	k := kid.New()

	k.Get("/greet/{name}", greetHandler)

	// It will be matched to any number of paramters. Including zero and one and so on.
	k.Get("/path/{*starParam}", starParamHandler)

	k.Run()
}

func greetHandler(c *kid.Context) {
	name := c.Param("name")

	c.String(http.StatusOK, fmt.Sprintf("Hello %s!", name))
}

func starParamHandler(c *kid.Context) {
	param := c.Param("starParam")

	c.String(http.StatusOK, fmt.Sprintf("Star path parameter: %q", param))
}
