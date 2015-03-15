package processPairController;

import(
	"os"
	"../systemController"
	"../log"
	"../network"
	"time"
	"../encoder/JSON"
);

//-----------------------------------------------//

const(
	ALIVE_MESSAGE_DEADLINE  		= 500
	ALIVE_NOTIFICATION_DELAY  		= 25
);

//-----------------------------------------------//

func backup() {

	aliveReceiver := make(chan string);
	dataReceiver  := make(chan string);

	go network.ListenWithDeadline(	network.GetNewAddress("localhost", 9871), aliveReceiver, time.Millisecond * ALIVE_MESSAGE_DEADLINE);
	go network.Listen(				network.GetNewAddress("localhost", 9872), dataReceiver);

	for {
		select {
			case aliveMessage := <- aliveReceiver:
				log.Data("Alive", aliveMessage);
			case data 		  := <- dataReceiver:
				log.Error("Got data", data);
		}
	}
}

//-----------------------------------------------//

func aliveNotification(aliveChannel chan string) {
	
	for {
		time.Sleep(time.Millisecond * ALIVE_NOTIFICATION_DELAY);
		
		message, _ := JSON.Encode("Alive");
		aliveChannel <- string(message);
	}
}

func master() {

	aliveChannel := make(chan string);
	go network.Send(network.GetNewAddress("localhost", 9871), aliveChannel);
	go aliveNotification(aliveChannel);

	go systemController.Run();
}

//-----------------------------------------------//

func Run() {

	if len(os.Args) >= 2 && os.Args[1] == "backup" {

		go backup();

	} else {

		go master();

	}
}