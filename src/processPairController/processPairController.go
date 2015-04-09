package processPairController;

import(
	"os"
	"../elevatorController"
	"../log"
	"../network"
	"time"
	"../encoder/JSON"
);

//-----------------------------------------------//

const(
	ALIVE_MESSAGE_DEADLINE  		= 200
	ALIVE_NOTIFICATION_DELAY  		= 15
);

//-----------------------------------------------//

func backupProcess() {

	addServerRecipientChannel := make(chan network.Recipient);

	aliveRecipient := network.Recipient{ Name : "alive", Channel : make(chan []byte) };

	timeoutTriggerTime 	:= time.Millisecond * ALIVE_MESSAGE_DEADLINE;
	timeoutNotifier 	:= make(chan bool);

	go network.ListenServerWithTimeout("localhost", 9132, addServerRecipientChannel, timeoutTriggerTime, timeoutNotifier);

	addServerRecipientChannel <- aliveRecipient;

	loop:
	for {
		select {
			case aliveMessage := <- aliveRecipient.Channel:
				log.Data("Alive", aliveMessage);
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

	go network.TransmitServer("localhost", 9132, aliveTransmitChannel);

	for {
		time.Sleep(time.Millisecond * ALIVE_NOTIFICATION_DELAY);
		aliveMessage, _ := JSON.Encode("Alive");
		aliveTransmitChannel <- network.Message{ RecipientName : "alive", Data : aliveMessage };
	}
}

func masterProcess() {

	elevatorController.Initialize();

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