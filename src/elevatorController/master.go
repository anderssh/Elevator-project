package elevatorController;

import(
	//"../typeDefinitions"
	"../network"
	"../log"
	"time"
	"../encoder/JSON"
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

type costBid struct {
	Value			int
	SenderIPAddr 	string
}

var costBids []costBid;

func costBidAddAndSort(costBids []costBid, newCostBid costBid) []costBid{
	
	costBidAllreadyInSlice := false;

	for i := 0; i < len(costBids); i++ {
		if (costBids[i].SenderIPAddr == newCostBid.SenderIPAddr) {
			costBidAllreadyInSlice = true;
		}
	}
	if (!costBidAllreadyInSlice) {
		costBids = append(costBids, newCostBid);
		for costBidIndex := (len(costBids) - 1); costBidIndex > 0; costBidIndex--{
			
			tempCostBid := costBids[costBidIndex]

			if (costBids[costBidIndex].Value < costBids[costBidIndex-1].Value) {
				
				costBids[costBidIndex] 		= costBids[costBidIndex-1]
				costBids[costBidIndex-1] 	= tempCostBid;
			}
		}
	}
	return costBids;
}
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

			if err != nil{
				log.Error(err);
			}

			log.Data("Master: Got cost", cost, message.SenderIPAddr);

			newCostBid := costBid{ Value : cost, SenderIPAddr : message.SenderIPAddr }
			
			costBids = costBidAddAndSort(costBids, newCostBid);

			log.Data(costBids[0].Value);


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
		time.Sleep(time.Millisecond * 200);
	}
}

func masterHandleAliveNotification(message network.Message) {

	switch currentState {
		case STATE_IDLE:

			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				print("Merge")
				currentState = STATE_AWAITING_DATA_COLLECTION;

				//spamCollectData()
			}

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:

			// Reset timer


	}
}

func masterHandleAliveNotificationTimeout(message network.Message) {

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:



		case STATE_INACTIVE:

			// Make master

	}
}

//-----------------------------------------------//

func master(transmitChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	costBids = make([]costBid, 0, 1);

	newOrderRecipient 			:= network.Recipient{ ID : "masterNewOrder", 			ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 		:= network.Recipient{ ID : "masterCostResponse", 		ReceiveChannel : make(chan network.Message) };
	aliveNotificationRecipient 	:= network.Recipient{ ID : "masterAliveNotification", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	addServerRecipientChannel <- aliveNotificationRecipient;
	
	currentState = STATE_IDLE;

	go masterAliveNotifier(transmitChannel);

	for {
		select {
			case message := <- newOrderRecipient.ReceiveChannel:

				masterHandleEventNewOrder(message, transmitChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				masterHandleEventCostResponse(message, transmitChannel);

			case message := <- aliveNotificationRecipient.ReceiveChannel:

				masterHandleAliveNotification(message);
		}	
	}
}
