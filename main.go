package main;

import (
	"user/processPairController"
	"runtime"
);

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	
	processPairController.Run();

	d_chan := make(chan bool, 1);
	<-d_chan;
}
