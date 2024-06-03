package main

import (
	"devilginx/server"
)

func main() {
	r := server.GetRouter()

	r.RunTLS("localhost:8443", "./cmd/devilginx-poc/testcert.crt", "./cmd/devilginx-poc/test.key")
}
