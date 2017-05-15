package main

import (
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
)

func InitWeb() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		// TODO: fetch drops

		c.HTML(http.StatusOK, "index.html.tmpl", []Drop{{"aaaa"}, {"bbbb"}})
	})
	r.POST("/", func(c *gin.Context) {
		text := c.PostForm("text")
		if text == "" {
			c.String(http.StatusBadRequest, "empty text")
			return
		}

		// TODO: save drop

		c.String(http.StatusCreated, "OK")
	})
	r.Run()
}
