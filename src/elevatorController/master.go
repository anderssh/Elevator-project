package elevatorController;

import(
	//"../typeDefinitions"
	"../network"
	"../config"
	"../log"
	"time"
	"../encoder/JSON"
	"strings"
	"strconv"
);

//-----------------------------------------------//

type State int

type costBid struct {
	Value			int
	SenderIPAddr 	string
}

var costBids []costBid;

const (
	STATE_IDLE   								State = iota
	STATE_AWAITING_COST_RESPONSE   				State = iota
	STATE_AWAITING_ORDER_TAKEN_CONFIRMATION		State = iota
	STATE_AWAITING_DATA_COLLECTION 				State = iota
	STATE_INACTIVE 								State = iota
);

var currentState State;

//-----------------------------------------------//

func masterHandleEventNewOrder(message network.Message, transmitChannel chan network.Message) {
	
	orderEncoded := message.Data;
	
	switch currentState {
		case STATE_IDLE:
			log.Data("Master: Got new order to distribute")
			transmitChannel <- network.MakeMessage("slaveCostRequest", orderEncoded, network.BROADCAST_ADDR);

			currentState = STATE_AWAITING_COST_RESPONSE;

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

func masterHandleEventCostResponse(message network.Message, transmitChannel chan network.Message){

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

			var cost int;
			err := JSON.Decode(message.Data, &cost);

			log.Error(err);
			log.Data("Master: Got cost", cost, message.SenderIPAddr);
			newCostBid := costBid{ Value : cost, SenderIPAddr : message.SenderIPAddr }
			costBids = append(costBids, newCostBid);

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

//-----------------------------------------------//

func masterAliveNotifier(transmitChannel chan network.Message) {

	for {
		messageContent, _ := JSON.Encode("Master alive");
		transmitChannel <- network.MakeMessage("masterAliveNotification", messageContent, network.BROADCAST_ADDR);
		time.Sleep(config.MASTER_ALIVE_NOTIFICATION_DELAY);
	}
}

func masterHandleAliveNotification(message network.Message, masterAliveTimeout *time.Timer) {

	switch currentState {
		case STATE_IDLE:

			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				print("Merge")
				currentState = STATE_AWAITING_DATA_COLLECTION;
			}

		case STATE_INACTIVE:

			masterAliveTimeout.Reset(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT);
	}
}

func masterHandleMasterDisconnect(masterAliveTimeout *time.Timer) {

	switch currentState {
		case STATE_INACTIVE:

			log.Data("No master, I return from the dead.");
			masterAliveTimeout.Stop();
			currentState = STATE_IDLE; 		// Make master
	}
}

//-----------------------------------------------//

var slaveDisconnectNotifiers map[string]*time.Timer = make(map[string]*time.Timer);

func masterDisplaySlaves() {

	log.DataWithColor(log.COLOR_RED, "Slaves:");

	for slaveIP, _ := range slaveDisconnectNotifiers {
		log.Data(slaveIP);
	}
}

func masterHandleSlaveAliveNotification(message network.Message, eventSlaveDisconnect chan string) {

	switch currentState {
		case STATE_IDLE:

			_, keyExists := slaveDisconnectNotifiers[message.SenderIPAddr];

			if keyExists {
				
				slaveDisconnectNotifiers[message.SenderIPAddr].Reset(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT);
				log.Data("Slave in list, reset...");

			} else {

				slaveDisconnectNotifiers[message.SenderIPAddr] = time.AfterFunc(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT, func() {
					eventSlaveDisconnect <- message.SenderIPAddr;
				});
				log.Data("Slave not in list, add...");
			}

		default:

			_, keyExists := slaveDisconnectNotifiers[message.SenderIPAddr];

			if keyExists {
				
				slaveDisconnectNotifiers[message.SenderIPAddr].Reset(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT);
				log.Data("Slave in list, reset...");
			}
	}
}

func masterHandleSlaveDisconnect(slaveDisconnectIP string) {

	log.Data("Disconnected slave", slaveDisconnectIP);
	delete(slaveDisconnectNotifiers, slaveDisconnectIP);
}

//-----------------------------------------------//

func master(transmitChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	currentState = STATE_IDLE;
	costBids = make([]costBid, 0, 1);

	//-----------------------------------------------//

	newOrderRecipient 			:= network.Recipient{ ID : "masterNewOrder", 			ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 		:= network.Recipient{ ID : "masterCostResponse", 		ReceiveChannel : make(chan network.Message) };
	aliveNotificationRecipient 	:= network.Recipient{ ID : "masterAliveNotification", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	addServerRecipientChannel <- aliveNotificationRecipient;

	slaveAliveNotificationRecipient := network.Recipient{ ID : "masterSlaveAliveNotification", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- slaveAliveNotificationRecipient;
	
	//-----------------------------------------------//

	eventSlaveDisconnect := make(chan string);
	eventMasterDisconnect := make(chan bool);

	masterAliveTimeout := time.AfterFunc(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT, func() {
		eventMasterDisconnect <- true;
	});

	//-----------------------------------------------//

	go masterAliveNotifier(transmitChannel);

	for {
		select {

			//-----------------------------------------------//
			// Distribute order

			case message := <- newOrderRecipient.ReceiveChannel:

				masterHandleEventNewOrder(message, transmitChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				masterHandleEventCostResponse(message, transmitChannel);

			//-----------------------------------------------//
			// Master switching

			case message := <- aliveNotificationRecipient.ReceiveChannel:

				masterHandleAliveNotification(message, masterAliveTimeout);

			case  <- eventMasterDisconnect:

				masterHandleMasterDisconnect(masterAliveTimeout);

			//-----------------------------------------------//
			// Slave registration

			case message := <- slaveAliveNotificationRecipient.ReceiveChannel:

				masterHandleSlaveAliveNotification(message, eventSlaveDisconnect);

			case slaveDisconnectIP := <- eventSlaveDisconnect:

				masterHandleSlaveDisconnect(slaveDisconnectIP);
		}	
	}
}