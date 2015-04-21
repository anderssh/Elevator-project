package processPairController;

import(
	"os"
	"user/config"
	"user/elevatorController"
	"user/log"
	"user/network"
	"time"
	"user/encoder/JSON"
);

//-----------------------------------------------//

func backupProcess() {

	log.Data("Backup process: starting...");

	addServerRecipientChannel := make(chan network.Recipient);

	aliveRecipient := network.Recipient{ ID : "backupProcessAlive", ReceiveChannel : make(chan network.Message) };
	dataRecipient  := network.Recipient{ ID : "backupProcessData", ReceiveChannel : make(chan network.Message) };
 
	timeoutNotifier 	:= make(chan bool);

	go network.UDPListenServerWithTimeout(network.LOCALHOST, addServerRecipientChannel, config.BACKUP_PROCESS_ALIVE_MESSAGE_DEADLINE, timeoutNotifier);

	addServerRecipientChannel <- aliveRecipient;
	addServerRecipientChannel <- dataRecipient;

	loop:
	for {
		select {
			case message := <- aliveRecipient.ReceiveChannel:
				log.Data("backupProcessAlive", message);
			case message := <- dataRecipient.ReceiveChannel:
				log.Data("dataRecipient", message);
			case 			     <- timeoutNotifier:
				log.Warning("Switching to master process");

				go masterProcess();
				break loop;
		}
	}
}

//-----------------------------------------------//

func masterProcessAliveNotification(transmitChannelUDP chan network.Message) {
	
	for {
		time.Sleep(config.BACKUP_PROCESS_ALIVE_NOTIFICATION_SLEEP);
		aliveMessage, _ := JSON.Encode("backupProcessAlive");

		transmitChannelUDP <- network.MakeTimeoutMessage("backupProcessAlive", aliveMessage, network.LOCALHOST);
	}
}

func masterProcess() {

	log.Data("Master process: starting...");

	transmitChannelUDP := make(chan network.Message);

	go network.UDPTransmitServer(transmitChannelUDP);

	go masterProcessAliveNotification(transmitChannelUDP);

	elevatorController.Initialize();
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