package main

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"net/http"
)

func initWeb(redisClient *redis.Client) {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", func(c *gin.Context) {
		drops, err := fetchAllDrops(redisClient)
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

		err := saveDrop(redisClient, text)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.Redirect(http.StatusSeeOther, "/")
	})
	r.Run()
}
