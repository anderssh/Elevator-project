package processPair;

import(
	"os/exec"
	. "user/typeDefinitions"
	"user/config"
	"user/distributionSystem"
	"user/log"
	"user/network"
	"time"
	"user/encoder/JSON"
	"io/ioutil"
);

//-----------------------------------------------//

func readBackupDataOrdersLocalFromFile() []OrderLocal {

	backupDataOrdersLocalEncoded, err := ioutil.ReadFile(config.BACKUP_FILE_NAME);

	if err != nil { 			// Did not open file, most likely not created yet
		log.Error(err);
		return make([]OrderLocal, 0, 1);
	}

	var backupDataOrdersLocal []OrderLocal;
	err = JSON.Decode(backupDataOrdersLocalEncoded, &backupDataOrdersLocal);

	if err != nil { 			// Corrupt or empty file
		log.Error(err);
		return make([]OrderLocal, 0, 1);	
	}

	return backupDataOrdersLocal;
}

//-----------------------------------------------//

func secondaryProcess() {

	log.Data("Secondary process: starting...");

	timeoutNotifier := make(chan bool);

	addServerRecipientChannel := make(chan network.Recipient);

	go network.UDPListenServerWithTimeout(network.LOCALHOST, addServerRecipientChannel, config.BACKUP_PROCESS_ALIVE_MESSAGE_DEADLINE, timeoutNotifier);

	aliveRecipient := network.Recipient{ ID : "secondaryProcessAlive", ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- aliveRecipient;

	loop:
	for {
		select {
			case <- aliveRecipient.ReceiveChannel:

				// Alive
			
			case <- timeoutNotifier:

				log.Warning("Backup process: switching to master process");

				go primaryProcess();
				break loop;
		}
	}
}

//-----------------------------------------------//

func primaryProcessAliveNotification(transmitChannelUDP chan network.Message) {
	
	for {
		time.Sleep(config.BACKUP_PROCESS_ALIVE_NOTIFICATION_SLEEP);
		aliveMessage, _ := JSON.Encode("Alive");

		transmitChannelUDP <- network.MakeTimeoutServerMessage("secondaryProcessAlive", aliveMessage, network.LOCALHOST);
	}
}

func primaryProcess() {

	log.Data("Primary process: starting...");

	transmitChannelUDP := make(chan network.Message);

	go network.UDPTransmitServer(transmitChannelUDP);

	//-----------------------------------------------//

	if config.SHOULD_USE_PROCESS_PAIRS {

		cmd := exec.Command("gnome-terminal", "-e", "./main");
		cmd.Output();

		log.Data("Primary process: Spawned backup");

		go primaryProcessAliveNotification(transmitChannelUDP);
	}

	//-----------------------------------------------//

	go distributionSystem.Run(transmitChannelUDP, readBackupDataOrdersLocalFromFile());
}

//-----------------------------------------------//

func Run() {

	network.Initialize();

	go secondaryProcess();
}