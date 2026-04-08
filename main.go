package main

import (
	"missfit/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
