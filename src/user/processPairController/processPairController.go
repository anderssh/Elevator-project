package processPairController;

import(
	"os"
	"user/elevatorController"
	"user/log"
	"user/network"
	"time"
	"user/encoder/JSON"
);

//-----------------------------------------------//

const(
	ALIVE_MESSAGE_DEADLINE  		= 200
	ALIVE_NOTIFICATION_DELAY  		= 15
);

//-----------------------------------------------//

func backupProcess() {

	addServerRecipientChannel := make(chan network.Recipient);

	aliveRecipient := network.Recipient{ ID : "alive", ReceiveChannel : make(chan network.Message) };

	timeoutTriggerTime 	:= time.Millisecond * ALIVE_MESSAGE_DEADLINE;
	timeoutNotifier 	:= make(chan bool);

	go network.UDPListenServerWithTimeout(network.LOCALHOST, addServerRecipientChannel, timeoutTriggerTime, timeoutNotifier);

	addServerRecipientChannel <- aliveRecipient;

	loop:
	for {
		select {
			case message := <- aliveRecipient.ReceiveChannel:
				log.Data("Alive", message);
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

	go network.UDPTransmitServer(aliveTransmitChannel);

	for {
		time.Sleep(time.Millisecond * ALIVE_NOTIFICATION_DELAY);
		aliveMessage, _ := JSON.Encode("Alive");
		aliveTransmitChannel <- network.MakeTimeoutMessage("alive", aliveMessage, network.LOCALHOST);
	}
}

func masterProcess() {

	elevatorController.Initialize();

	//go masterProcessAliveNotification();
	go elevatorController.Run();
}

//-----------------------------------------------//

func Run() {

	network.Initialize();

	if len(os.Args) >= 2 && os.Args[1] == "backup" {

		go backupProcess();

	} else {

		go masterProcess();

	}
}