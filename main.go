package main

import (
	"./src/systemController"
)

func main() {

	systemController.Run()

	d_chan := make(chan bool, 1)
	<-d_chan
}
