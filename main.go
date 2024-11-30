package main

import (
	"feed/initializers"
	"feed/routes"
)

func init() {
	initializers.LoadEnvVar()
	initializers.ConnectDB()
	initializers.OpenRedis()
}
func main() {

	r := routes.SetupRoutes()
	r.Run()
}
