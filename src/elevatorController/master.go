package elevatorController;

import(
	//"../typeDefinitions"
	"../network"
	"../log"
	"../encoder/JSON"
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

var currentState State = STATE_IDLE;

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

var 
func masterHandleEventCostResponse(message network.Message, transmitChannel chan network.Message){

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

			var cost int;
			err := JSON.Decode(message.Data, &cost);

			log.Error(err);
			log.Data("Master: Got cost", cost, message.SenderIPAddr);
			newCostBid := {Value: = cost, SenderIPAddr, message.SenderIPAddr}
			costBids = append(costBids, {})

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

		case STATE_AWAITING_DATA_COLLECTION:

		case STATE_INACTIVE:
	}
}

//-----------------------------------------------//

func master(transmitChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	newOrderRecipient 		:= network.Recipient{ ID : "masterNewOrder", 		ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 	:= network.Recipient{ ID : "masterCostResponse", 	ReceiveChannel : make(chan network.Message) };
	
	costBids = make([]costBid,0,1);

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	
	for {
		select {
			case message := <- newOrderRecipient.ReceiveChannel:

				masterHandleEventNewOrder(message, transmitChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				masterHandleEventCostResponse(message, transmitChannel);
		}	
	}
}
