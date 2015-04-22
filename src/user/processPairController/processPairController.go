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
	"io/ioutil"
);

//-----------------------------------------------//

func initializeBackupDataOrders() OrdersBackup {

	backupDataOrdersEncoded, err := ioutil.ReadFile(config.BACKUP_FILE_NAME);

	if err != nil { 			// Did not open file, most likely not created yet
		log.Error(err);
		return OrdersBackup{ Orders : make([]Order, 0, 1) };
	}

	var backupDataOrders OrdersBackup;
	err = JSON.Decode(backupDataOrdersEncoded, &backupDataOrders);

	if err != nil { 			// Corrupt or empty file
		log.Error(err);
		return OrdersBackup{ Orders : make([]Order, 0, 1) };	
	}

	return backupDataOrders;
}

func writeBackupDataOrdersToFile(writeBackupDataOrders chan []byte) {

	for {
		dataToWrite := <- writeBackupDataOrders;

		ioutil.WriteFile(config.BACKUP_FILE_NAME, dataToWrite, 0644); // 0644: for read/write permissions
	}
}

func initializeBackupDataOrdersGlobal() OrdersGlobalBackup {
	return OrdersGlobalBackup{ Orders : make([]OrderGlobal, 0, 1) };
}

//-----------------------------------------------//

func handleBackupDataOrders(message network.Message, backupDataOrders OrdersBackup, writeBackupDataOrders chan []byte) OrdersBackup {

	var dataReceived OrdersBackup;
	err := JSON.Decode(message.Data, &dataReceived);

	if err != nil {
		log.Error(err);
	}

	if dataReceived.Timestamp >= backupDataOrders.Timestamp {
		
		log.Data("Backup process: new backup data destination orders received.");
		
		writeBackupDataOrders <- message.Data;

		return dataReceived;
	} else {
		return backupDataOrders;
	}
}

func handleBackupDataOrdersGlobal(message network.Message, backupDataOrdersGlobal OrdersGlobalBackup) OrdersGlobalBackup {
	
	var dataReceived OrdersGlobalBackup;
	err := JSON.Decode(message.Data, &dataReceived);

	if err != nil {
		log.Error(err);
	}

	if dataReceived.Timestamp >= backupDataOrdersGlobal.Timestamp {
		log.Data("Backup process: new backup data orders global received.");
		return dataReceived;
	} else {
		return backupDataOrdersGlobal;
	}
}

//-----------------------------------------------//

func backupProcess() {

	log.Data("Backup process: starting...");

	writeBackupDataOrders := make(chan []byte, 1000);

	go writeBackupDataOrdersToFile(writeBackupDataOrders);

	backupDataOrders 		:= initializeBackupDataOrders();
	backupDataOrdersGlobal 	:= initializeBackupDataOrdersGlobal();

	addServerRecipientChannel := make(chan network.Recipient);

	timeoutNotifier 	:= make(chan bool);

	go network.UDPListenServerWithTimeout(network.LOCALHOST, addServerRecipientChannel, config.BACKUP_PROCESS_ALIVE_MESSAGE_DEADLINE, timeoutNotifier);

	aliveRecipient 					 := network.Recipient{ ID : "backupProcessAlive", ReceiveChannel : make(chan network.Message) };
	backupDataOrdersRecipient  		 := network.Recipient{ ID : "backupProcessDataOrders", ReceiveChannel : make(chan network.Message) };
	backupDataOrdersGlobalRecipient  := network.Recipient{ ID : "backupProcessDataOrdersGlobal", ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- aliveRecipient;
	addServerRecipientChannel <- backupDataOrdersRecipient;
	addServerRecipientChannel <- backupDataOrdersGlobalRecipient;

	loop:
	for {
		select {
			case <- aliveRecipient.ReceiveChannel:

				// Alive
				
			case message := <- backupDataOrdersRecipient.ReceiveChannel:
				
				backupDataOrders = handleBackupDataOrders(message, backupDataOrders, writeBackupDataOrders);

			case message := <- backupDataOrdersGlobalRecipient.ReceiveChannel:

				backupDataOrdersGlobal = handleBackupDataOrdersGlobal(message, backupDataOrdersGlobal);

			case <- timeoutNotifier:

				log.Warning("Backup process: switching to master process");

				go masterProcess(backupDataOrders, backupDataOrdersGlobal);
				break loop;
		}
	}
}

//-----------------------------------------------//

func masterProcessAliveNotification(transmitChannelUDP chan network.Message) {
	
	for {
		time.Sleep(config.BACKUP_PROCESS_ALIVE_NOTIFICATION_SLEEP);
		aliveMessage, _ := JSON.Encode("backupProcessAlive");

		transmitChannelUDP <- network.MakeTimeoutServerMessage("backupProcessAlive", aliveMessage, network.LOCALHOST);
	}
}

func masterProcess(backupDataOrders OrdersBackup, backupDataOrdersGlobal OrdersGlobalBackup) {

	log.Data("Master process: starting...");

	if config.SHOULD_USE_PROCESS_PAIRS {

		cmd := exec.Command("gnome-terminal", "-e", "./main");
		cmd.Output();

		log.Data("Master process: Spawned backup");
	}

	transmitChannelUDP := make(chan network.Message);

	go network.UDPTransmitServer(transmitChannelUDP);

	go masterProcessAliveNotification(transmitChannelUDP);

	go elevatorController.Run(transmitChannelUDP, backupDataOrders, backupDataOrdersGlobal);
}

//-----------------------------------------------//

func Run() {

	network.Initialize();

	go backupProcess();
}