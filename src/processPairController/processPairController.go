package processPairController;

import(
	"os"
	"../elevatorController"
	"../log"
	"../network"
	"time"
);

//-----------------------------------------------//

const(
	ALIVE_MESSAGE_DEADLINE  		= 200
	ALIVE_NOTIFICATION_DELAY  		= 15
);

//-----------------------------------------------//

func backupProcess() {

	addServerRecipientChannel := make(chan network.Recipient);

	aliveRecipient := network.Recipient{ Name : "alive", Channel : make(chan string) };
	dataRecipient  := network.Recipient{ Name : "data", Channel : make(chan string) };

	timeoutTriggerTime 	:= time.Millisecond * ALIVE_MESSAGE_DEADLINE;
	timeoutNotifier 	:= make(chan bool);

	go network.ListenServerWithTimeout("localhost", 10005, addServerRecipientChannel, timeoutTriggerTime, timeoutNotifier);

	addServerRecipientChannel <- aliveRecipient;
	addServerRecipientChannel <- dataRecipient;

	loop:
	for {
		select {
			case aliveMessage := <- aliveRecipient.Channel:
				log.Data("Alive", aliveMessage);
			case data 		  := <- dataRecipient.Channel:
				log.Data("Got data", data);
			case 			     <- timeoutNotifier:
				log.Warning("Switching to master process");

				go masterProcess();
				break loop;
		}
	}
}

//-----------------------------------------------//

func masterProcessAliveNotification() {
	
	aliveTransmitChannel := make(chan network.Message);

	go network.TransmitServer("localhost", 10005, aliveTransmitChannel);

	for {
		time.Sleep(time.Millisecond * ALIVE_NOTIFICATION_DELAY);
		aliveTransmitChannel <- network.Message{ Recipient : "alive", Data : "Alive" };
	}
}

func masterProcess() {

	go masterProcessAliveNotification();
	go elevatorController.Run();
}

//-----------------------------------------------//

func Run() {

	if len(os.Args) >= 2 && os.Args[1] == "backup" {

		go backupProcess();

	} else {

		go masterProcess();

	}
}