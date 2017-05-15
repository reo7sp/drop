package main

import (
	"gopkg.in/gin-gonic/gin.v1"
	"net/http"
	"github.com/go-redis/redis"
)

func InitWeb(redisClient *redis.Client) {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		drops, err := FetchAllDrops(redisClient)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.HTML(http.StatusOK, "index.html.tmpl", drops)
	})
	r.POST("/", func(c *gin.Context) {
		text := c.PostForm("text")
		if text == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		err := SaveDrop(redisClient, text)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusSeeOther, "/")
	})
	r.Run()
}
