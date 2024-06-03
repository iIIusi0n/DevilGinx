package main

import (
	"devilginx/server"
)

func main() {
	r := server.GetRouter()

	r.Run(":8080")
}
