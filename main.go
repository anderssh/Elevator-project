package main;

import (
	"./src/processPairController"
);

func main() {
	
	processPairController.Run();

	d_chan := make(chan bool, 1);
	<-d_chan;
}
