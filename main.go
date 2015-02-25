package main;

import(
	"./src/elevator"
);


func main() {


	s := elevator.Initialize();
	if (s) {

	}


	elevator.Run();

	d_chan := make(chan bool, 1);
	<- d_chan;
}