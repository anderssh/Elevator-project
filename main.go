package main

import (
	"./src/elevatorController"
)

func main() {

	elevatorController.Run()

	d_chan := make(chan bool, 1)
	<-d_chan

}
