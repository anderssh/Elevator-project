package main;

import (
	"user/processPair"
	"runtime"
);

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	
	go processPair.Run();

	d_chan := make(chan bool, 1);
	<-d_chan;
}
