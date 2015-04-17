package elevatorController;

import(
	. "user/typeDefinitions"
	"user/network"
	"user/config"
	"user/log"
	"time"
	"user/encoder/JSON"
	"strings"
	"strconv"
);

//-----------------------------------------------//

type State int

const (
	STATE_IDLE   								State = iota
	STATE_AWAITING_COST_RESPONSE   				State = iota
	STATE_AWAITING_ORDER_TAKEN_CONFIRMATION		State = iota
	STATE_AWAITING_DATA_COLLECTION 				State = iota
	STATE_INACTIVE 								State = iota
);

var currentState State;

//-----------------------------------------------//

var workerIPAddrs []string;

//-----------------------------------------------//

type CostBid struct {
	Value			int
	SenderIPAddr 	string
}

var costBids []CostBid;

func costBidAllreadyStored(costBid CostBid) bool {
		
	for costBidIndex := 0 ; costBidIndex < len(costBids); costBidIndex++{
		if (costBids[costBidIndex].SenderIPAddr == costBid.SenderIPAddr) {
			return true;
		}
	}
	
	return false;
}

func costBidAddAndSort(newCostBid CostBid) {
	
	costBids = append(costBids, newCostBid);
		
	for costBidIndex := (len(costBids) - 1); costBidIndex > 0; costBidIndex--{
		
		tempCostBid := costBids[costBidIndex]

		if (costBids[costBidIndex].Value < costBids[costBidIndex-1].Value) {
			
			costBids[costBidIndex] 		= costBids[costBidIndex-1];
			costBids[costBidIndex-1] 	= tempCostBid;
		}
	}
}

//-----------------------------------------------//
// Order handling

var currentlyHandledOrder Order;

func masterHandleEventNewOrder(message network.Message, transmitChannel chan network.Message) {
	
	switch currentState {
		case STATE_IDLE:
			
			log.Data("Master: Got new order to distribute")

			orderEncoded := message.Data;
			err := JSON.Decode(orderEncoded, &currentlyHandledOrder);

			if err != nil {
				log.Error(err, "decode error");
			}

			for worker := range workerIPAddrs {
				transmitChannel <- network.MakeMessage("slaveCostRequest", orderEncoded, workerIPAddrs[worker]);
			}

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
			log.DataWithColor(log.COLOR_YELLOW, "state idle")
		case STATE_AWAITING_COST_RESPONSE:

			var cost int;
			err := JSON.Decode(message.Data, &cost);

			if err != nil {
				log.Error(err);
			}

			log.Data("Master: Got cost", cost, "from", message.SenderIPAddr);

			newCostBid := CostBid{ Value : cost, SenderIPAddr : message.SenderIPAddr };

			if !costBidAllreadyStored(newCostBid) {
				costBidAddAndSort(newCostBid);
			}

			if len(workerIPAddrs) == len(costBids) {

				log.Data("Master: send destination", currentlyHandledOrder.Floor, "to", costBids[0].SenderIPAddr);
				
				order, _ := JSON.Encode(currentlyHandledOrder);
				transmitChannel <- network.MakeMessage("slaveNewDestinationOrder", order, costBids[0].SenderIPAddr);

				currentState = STATE_IDLE;
			}

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

//-----------------------------------------------//

func masterHandleEventOrderTakenConfirmation(message network.Message, transmitChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

			var takenOrder Order;

			err := JSON.Decode(message.Data, &takenOrder);

			if err != nil{
				log.Error(err, "Decode error");
			}

			log.Data("Master: Got order taken Confirmation")

			
		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

//-----------------------------------------------//

func masterHandleActiveNotificationTick(broadcastChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:

			message, _ := JSON.Encode("Alive");
			broadcastChannel <- network.MakeMessage("masterActiveNotification", message, network.BROADCAST_ADDR);
	}
}

func masterHandleActiveNotification(message network.Message, timeoutMasterActiveNotification *time.Timer, transmitChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:
			
			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				log.Data("Master: Merge with", message.SenderIPAddr);
				messageMerge, _ := JSON.Encode("Merge");
				transmitChannel <- network.MakeMessage("masterMergeRequest", messageMerge, message.SenderIPAddr);

				workerIPAddrs = append(workerIPAddrs, message.SenderIPAddr);

				currentState = STATE_IDLE;
			}

		case STATE_INACTIVE:

			timeoutMasterActiveNotification.Reset(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT);
	}
}

func masterHandleMergeRequest(message network.Message, eventChangeMaster chan string) {

	switch currentState {
		case STATE_IDLE:

			log.Data("Master: Going into inactive some else is my master now.");
			eventChangeMaster <- message.SenderIPAddr;

			currentState = STATE_INACTIVE;
	}
}

func masterHandleMasterDisconnect(timeoutMasterActiveNotification *time.Timer, eventInactiveDisconnect chan string, eventChangeNotificationRecipientID chan string) {

	switch currentState {
		case STATE_INACTIVE:

			log.Data("No master, switch");

			timeoutMasterActiveNotification.Stop();
			
			workerIPAddrs = make([]string, 0, 1);
			workerIPAddrs = append(workerIPAddrs, network.GetLocalIPAddr());

			eventChangeNotificationRecipientID <- "masterActiveNotification";

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//

func masterDisplayWorkers() {

	if config.SHOULD_DISPLAY_WORKERS {

		log.DataWithColor(log.COLOR_CYAN, "-------------------");
		log.DataWithColor(log.COLOR_CYAN, "Workers:");

		for worker := range workerIPAddrs {
			log.Data(workerIPAddrs[worker]);
		}

		log.DataWithColor(log.COLOR_CYAN, "-------------------");
	}
}

func masterHandleInactiveNotification(message network.Message) {

	/*_, keyExists := inactiveDisconnectTimeouts[message.SenderIPAddr];

	if keyExists {
			
		inactiveDisconnectTimeouts[message.SenderIPAddr].Reset(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT);
		log.Data("Inactive in list, reset...");
	}*/
}

func masterHandleInactiveDisconnect(slaveDisconnectIP string) {

	// Redistribute order of disconnected node
	log.Data("Disconnected slave", slaveDisconnectIP);
}

//-----------------------------------------------//

func master(transmitChannel 				chan network.Message,
			addServerRecipientChannel 		chan network.Recipient,
			broadcastChannel 				chan network.Message,
			addBroadcastRecipientChannel 	chan network.Recipient,
			eventChangeMaster 				chan string) {

	currentState = STATE_IDLE;

	workerIPAddrs = make([]string, 0, 1);
	workerIPAddrs = append(workerIPAddrs, network.GetLocalIPAddr());

	costBids = make([]CostBid, 0, 1);

	//-----------------------------------------------//

	activeNotificationRecipient := network.Recipient{ ID : "masterActiveNotification", 		ReceiveChannel : make(chan network.Message) };

	addBroadcastRecipientChannel <- activeNotificationRecipient;

	eventActiveNotificationTicker := time.NewTicker(config.MASTER_ALIVE_NOTIFICATION_DELAY);

	//-----------------------------------------------//

	newOrderRecipient 				:= network.Recipient{ ID : "masterNewOrder", 				ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 			:= network.Recipient{ ID : "masterCostResponse", 			ReceiveChannel : make(chan network.Message) };
	orderTakenConfirmationRecipient := network.Recipient{ ID : "masterOrderTakenConfirmation", 	ReceiveChannel : make(chan network.Message) };

	mergeRequestRecipient 			:= network.Recipient{ ID : "masterMergeRequest", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	addServerRecipientChannel <- orderTakenConfirmationRecipient;

	addServerRecipientChannel <- mergeRequestRecipient;

	eventInactiveDisconnect 			:= make(chan string);
	eventActiveNotificationTimeout 		:= make(chan bool);
	eventChangeNotificationRecipientID 	:= make(chan string);

	timeoutMasterActiveNotification := time.AfterFunc(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT, func() {
		//eventActiveNotificationTimeout <- true;
	});
	
	//-----------------------------------------------//

	for {
		select {

			//-----------------------------------------------//
			// Distribute order

			case message := <- newOrderRecipient.ReceiveChannel:

				masterDisplayWorkers();
				masterHandleEventNewOrder(message, transmitChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				masterHandleEventCostResponse(message, transmitChannel);

			case message := <- orderTakenConfirmationRecipient.ReceiveChannel:

				masterHandleEventOrderTakenConfirmation(message, transmitChannel);

			//-----------------------------------------------//
			// Master switching

			case <- eventActiveNotificationTicker.C:

				masterHandleActiveNotificationTick(broadcastChannel);

			case message := <- activeNotificationRecipient.ReceiveChannel:
				
				masterHandleActiveNotification(message, timeoutMasterActiveNotification, transmitChannel);

			case message := <- mergeRequestRecipient.ReceiveChannel:

				masterHandleMergeRequest(message, eventChangeMaster);

			case  <- eventActiveNotificationTimeout:

				masterHandleMasterDisconnect(timeoutMasterActiveNotification, eventInactiveDisconnect, eventChangeNotificationRecipientID);

			//-----------------------------------------------//
			// Inactive registration

			case inactiveDisconnectIP := <- eventInactiveDisconnect:

				masterHandleInactiveDisconnect(inactiveDisconnectIP);
		}	
	}
}