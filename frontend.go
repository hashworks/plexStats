package main

import (
	"github.com/gin-gonic/gin"

	"net/http"
)

func (s server) frontendHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.tmpl", gin.H{
		"title": "Foo",
		"foo":   "Bar",
	})
}
