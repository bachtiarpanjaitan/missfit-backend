package main

import (
	"lumos/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
