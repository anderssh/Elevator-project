package processPairController;

import(
	"os/exec"
	. "user/typeDefinitions"
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

	backupData := OrdersGlobalBackup{ Orders : make([]OrderGlobal, 0, 1) };

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
			case 			<- aliveRecipient.ReceiveChannel:
				// Alive
			case message := <- dataRecipient.ReceiveChannel:
				
				var backupDataReceived OrdersGlobalBackup;
				err := JSON.Decode(message.Data, &backupDataReceived);

				if err != nil {
					log.Error(err);
				}

				if backupDataReceived.Timestamp >= backupData.Timestamp {
					log.Data("Backup process: new backup data received.");
					backupData = backupDataReceived;
				}

			case 			     <- timeoutNotifier:
				log.Warning("Backup process: switching to master process");

				go masterProcess(backupData);
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

func masterProcess(backupData OrdersGlobalBackup) {

	log.Data("Master process: starting...");

	cmd := exec.Command("gnome-terminal", "-e", "./main");
	cmd.Output();

	log.Data("Master process: Spawned backup");

	transmitChannelUDP := make(chan network.Message);

	go network.UDPTransmitServer(transmitChannelUDP);

	go masterProcessAliveNotification(transmitChannelUDP);

	elevatorController.Initialize();
	go elevatorController.Run(transmitChannelUDP, backupData);
}

//-----------------------------------------------//

func Run() {

	network.Initialize();

	go backupProcess();
}