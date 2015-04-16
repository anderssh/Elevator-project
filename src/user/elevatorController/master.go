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

var inactiveDisconnectTimeouts map[string]*time.Timer;

//-----------------------------------------------//

type costBid struct {
	Value			int
	SenderIPAddr 	string
}

var costBids []costBid;

func costBidAllreadyStored(costBid Costbid){
		
		for costBidIndex := 0 ; costBidIndex < len(costBids); costBidIndex++{
			if (costBids[costBidIndex].SenderIPAddr == costBid.SenderIPAddr) {
				return true;
			}
		}
		
		return false;
}

func costBidAddAndSort(costBids []costBid, newCostBid costBid) []costBid{
	
	costBids = append(costBids, newCostBid);
		
	for costBidIndex := (len(costBids) - 1); costBidIndex > 0; costBidIndex--{
		
		tempCostBid := costBids[costBidIndex]

		if (costBids[costBidIndex].Value < costBids[costBidIndex-1].Value) {
			
			costBids[costBidIndex] 		= costBids[costBidIndex-1];
			costBids[costBidIndex-1] 	= tempCostBid;
		}
	}

	return costBids;
}

//-----------------------------------------------//
// Order handling

var currentlyHandledOrder Order;

func masterHandleEventNewOrder(message network.Message, transmitChannel chan network.Message, timeoutCostResponse *time.Timer) {
	
	switch currentState {
		case STATE_IDLE:
			
			log.Data("Master: Got new order to distribute")

			orderEncoded := message.Data;
			err := JSON.Decode(orderEncoded, &currentlyHandledOrder);

			if err != nil {
				log.Error(err);
			}

			network.Repeat(transmitChannel, network.MakeMessage("slaveCostRequest", orderEncoded, network.BROADCAST_ADDR), 8, 20);
			
			timeoutCostResponse.Reset(config.TIMEOUT_TIME_COST_RESPONSE);
			log.Error(timeoutCostResponse, "asd");

			currentState = STATE_AWAITING_COST_RESPONSE;

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

func masterHandleEventCostResponse(message network.Message, transmitChannel chan network.Message, timeoutCostResponse *time.Timer){

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

			var cost int;
			err := JSON.Decode(message.Data, &cost);

			if err != nil{
				log.Error(err);
			}

			log.Data("Master: Got cost", cost, message.SenderIPAddr);

			newCostBid := costBid{ Value : cost, SenderIPAddr : message.SenderIPAddr };

			if (!(costBidAllreadyStored(newCostBid))) {
				costBids = costBidAddAndSort(costBids, newCostBid);
			}

			if (len(inactiveDisconnectTimeouts) + 1 == len(costBids)) {

				log.Data("Send destination", currentlyHandledOrder.Floor, currentlyHandledOrder.Type);
				
				order, _ := JSON.Encode(currentlyHandledOrder);
				transmitChannel <- network.MakeMessage("slaveNewDestinationOrder", order, costBids[0].SenderIPAddr);

				timeoutCostResponse.Stop();

				currentState = STATE_IDLE;
			}

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

func masterHandleEventTimeoutCostResponse() {

	switch currentState {

		case STATE_AWAITING_COST_RESPONSE:

			log.Data("Timeout, all participants must be involved. Reset...");
			
			costBids = make([]costBid, 0, 1);

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//

func masterHandleEventOrderTakenConfirmation(message network.Message, transmitChannel chan network.Message, timeoutOrderTakenConfirmation *time.Timer) {

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

			var takenOrder order;

			err := JSON.Decode(message.Data, &takenOrder);

			if err != nil{
				log.Error(err);
			}

			log("master: Got order taken Confirmation")
		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

}
//-----------------------------------------------//

func masterHandleActiveNotification(message network.Message, timeoutMasterActiveNotification *time.Timer) {

	switch currentState {
		case STATE_IDLE:

			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				print("Merge");
				currentState = STATE_AWAITING_DATA_COLLECTION;
			}

		case STATE_INACTIVE:

			timeoutMasterActiveNotification.Reset(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT);
	}
}

func masterHandleMasterDisconnect(timeoutMasterActiveNotification *time.Timer, eventInactiveDisconnect chan string, eventChangeNotificationRecipientID chan string) {

	switch currentState {
		case STATE_INACTIVE:

			log.Data("No master, switch");

			timeoutMasterActiveNotification.Stop();
			inactiveDisconnectTimeouts = make(map[string]*time.Timer);

			eventChangeNotificationRecipientID <- "masterActiveNotification";

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//

func masterDisplayInactive() {

	if config.SHOULD_DISPLAY_SLAVES {

		log.DataWithColor(log.COLOR_CYAN, "Slaves:");
		log.Data(network.GetLocalIPAddr());

		for slaveIP, _ := range inactiveDisconnectTimeouts {
			log.Data(slaveIP);
		}

		log.DataWithColor(log.COLOR_CYAN, "-------------------");
	}
}

func masterHandleInactiveNotification(message network.Message, eventInactiveDisconnect chan string) {

	_, keyExists := inactiveDisconnectTimeouts[message.SenderIPAddr];

	if keyExists {
			
		inactiveDisconnectTimeouts[message.SenderIPAddr].Reset(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT);
		log.Data("Inactive in list, reset...");
	}
}

func masterHandleInactiveDisconnect(slaveDisconnectIP string) {

	// Redistribute order of disconnected node
	log.Data("Disconnected slave", slaveDisconnectIP);
	delete(inactiveDisconnectTimeouts, slaveDisconnectIP);
}

//-----------------------------------------------//

func aliveNotification(transmitChannel chan network.Message, eventChangeNotificationRecipientID chan string) {

	eventTick := time.NewTicker(config.MASTER_ALIVE_NOTIFICATION_DELAY);
	recipientID := "masterInactiveNotification"; // | masterActiveNotification

	for {
		select {
			case <- eventTick.C:
				
				message, _ := JSON.Encode("Alive");
				transmitChannel <- network.MakeMessage(recipientID, message, network.BROADCAST_ADDR);
			
			case newRecepientID := <- eventChangeNotificationRecipientID:
				
				recipientID = newRecepientID;			
		}
	}
}

//-----------------------------------------------//

func master(transmitChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	currentState = STATE_INACTIVE;

	inactiveDisconnectTimeouts = make(map[string]*time.Timer);

	costBids = make([]costBid, 0, 1);

	//-----------------------------------------------//

	newOrderRecipient 				:= network.Recipient{ ID : "masterNewOrder", 				ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 			:= network.Recipient{ ID : "masterCostResponse", 			ReceiveChannel : make(chan network.Message) };
	orderTakenConfirmationRecipient := network.Recipient{ ID : "masterOrderTakenConfirmation", 	ReceiveChannel : make(chan network.Message) };
	aliveNotificationRecipient 		:= network.Recipient{ ID : "masterActiveNotification", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	addServerRecipientChannel <- orderTakenConfirmationRecipient;
	addServerRecipientChannel <- aliveNotificationRecipient;


	inactiveNotificationRecipient := network.Recipient{ ID : "masterInactiveNotification", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- inactiveNotificationRecipient;
	
	//-----------------------------------------------//

	eventInactiveDisconnect 			:= make(chan string);
	eventActiveNotificationTimeout 		:= make(chan bool);
	eventChangeNotificationRecipientID 	:= make(chan string);

	timeoutMasterActiveNotification := time.AfterFunc(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT, func() {
		eventActiveNotificationTimeout <- true;
	});

	eventTimeoutCostReponse 			:= make(chan bool);
	
	timeoutCostResponse := time.AfterFunc(config.TIMEOUT_TIME_COST_RESPONSE, func() {
			eventTimeoutCostReponse <- true;
	});
	timeoutCostResponse.Stop();
	
	//-----------------------------------------------//

	go aliveNotification(transmitChannel, eventChangeNotificationRecipientID);

	//-----------------------------------------------//

	for {
		select {

			//-----------------------------------------------//
			// Distribute order

			case message := <- newOrderRecipient.ReceiveChannel:

				masterHandleEventNewOrder(message, transmitChannel, timeoutCostResponse);
				log.Error(timeoutCostResponse)
			
			case message := <- costResponseRecipient.ReceiveChannel:

				masterHandleEventCostResponse(message, transmitChannel, timeoutCostResponse);

			case <- eventTimeoutCostReponse:

				log.Warning("asd")
				masterHandleEventTimeoutCostResponse();

			case message := <- orderTakenConfirmationRecipient.ReceiveChannel:

				masterHandleEventOrderTakenConfirmation(message, transmitChannel, timeoutOrderTakenConfirmation)

			//-----------------------------------------------//
			// Master switching

			case message := <- aliveNotificationRecipient.ReceiveChannel:

				masterHandleActiveNotification(message, timeoutMasterActiveNotification);

			case  <- eventActiveNotificationTimeout:

				masterHandleMasterDisconnect(timeoutMasterActiveNotification, eventInactiveDisconnect, eventChangeNotificationRecipientID);

			//-----------------------------------------------//
			// Inactive registration

			case message := <- inactiveNotificationRecipient.ReceiveChannel:
				
				masterDisplayInactive();
				masterHandleInactiveNotification(message, eventInactiveDisconnect);

			case inactiveDisconnectIP := <- eventInactiveDisconnect:

				masterHandleInactiveDisconnect(inactiveDisconnectIP);
		}	
	}
}