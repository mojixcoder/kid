package main

import (
	"net/http"

	"github.com/mojixcoder/kid"
)

type MeRequest struct {
	LastName string `json:"last_name"`
}

type MeResponse struct {
	FullName string `json:"full_name"`
	Age      int    `json:"age"`
}

func main() {
	k := kid.New()

	k.Get("/me", meHandler)

	k.Run()
}

func meHandler(c *kid.Context) {
	var req MeRequest

	if err := c.ReadJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, kid.Map{"message": "invalid request body"})
		return
	}

	me := MeResponse{
		FullName: "Mojix " + req.LastName,
		Age:      23,
	}

	c.JSON(http.StatusOK, me)
}
