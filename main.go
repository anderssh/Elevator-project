package main;

import(
	"./src/network"
	"./src/driver"
	"./src/log"
	"./src/formatter/JSON"
	"time"
	"net"
	"fmt"
);

func read(r chan network.Message) {
	for {
		select {
			case <- r:
				
				//fmt.Println(message.Data);
		}
	}
}

func send(s chan network.Message) {
	for {
		time.Sleep(time.Second);
		add, _ := net.ResolveUDPAddr("udp", "129.241.187.143:20005");
		message := network.Message{ Length : 0, Data : "Fra 2", RemoteAddress : add };
		s <- message;
	}
}

func main() {

	fmt.Println("START:..");

	log.SetLogLevel(log.LOG_LEVEL_DEBUG);
	log.Debug(123);
	log.Error("Mayday, mayday!");

	type person struct{
		Skostrls 	int
		Navn 		string
	}

	var Patrick person;
	Patrick.Skostrls = 45;
	Patrick.Navn = "Patrish";

	hallo, _ := JSON.Encode(Patrick);
	fmt.Println(Patrick);

	fmt.Println(string(hallo));
	var halla person;
	error := JSON.Decode(hallo, &halla);
	
	fmt.Println(halla);
	fmt.Println(error);

	receiveChannel := make(chan network.Message);
	sendChannel := make(chan network.Message);

	network.Initialize(20005, receiveChannel, sendChannel);
	
	go read(receiveChannel);
	go send(sendChannel);

	s := driver.Initialize();
	fmt.Println(s);

	d_chan := make(chan bool, 1);
	<- d_chan;
}