package main

import (
	"devilginx/server"

	"github.com/gin-gonic/autotls"
)

func main() {
	r := server.GetRouter()

	autotls.Run(r, "localhost")
}
