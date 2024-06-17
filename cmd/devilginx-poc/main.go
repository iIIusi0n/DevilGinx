package main

import (
	"devilginx/server"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/autotls"
)

func main() {
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	log.SetOutput(io.MultiWriter(file, os.Stdout))

	r := server.GetRouter()

	go func() {
		redirector := gin.New()
		redirector.GET("/*path", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "https://localhost:8443")
		})
		redirector.Run(":8080")
	}()

	autotls.Run(r, "localhost")
}
