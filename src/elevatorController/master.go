package elevatorController;

import(
	. "../typeDefinitions"
	"../network"
	"../encoder/JSON"
);

//-----------------------------------------------//

type State int

const (
	STATE_IDLE   								State = iota
	STATE_AWAITING_COST_RESPONSE   				State = iota
	STATE_AWAITING_ORDER_TAKEN_CONFIRMATION		State = iota
	STATE_AWAITING_MASTER_DATA_COLLECTION 		State = iota
	STATE_INACTIVE 								State = iota
);

currentState State = STATE_IDLE;

//-----------------------------------------------//

func handleEventNewOrder(orderEncoded []byte) {
	
	switch currentState {
		case STATE_IDLE:

			broadcastChannel <- network.Message{ RecipientName : "slaveOrderRequest", Data : orderEncoded };

			currentState = STATE_AWAITING_COST_RESPONSE;

		case STATE_AWAITING_COST_RESPONSE:


	}
}

func master(broadcastChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	newOrderRecipient := network.Recipient{ Name : "masterNewOrder", Channel : make(chan []byte) };
	newCostResponseRecipient := network.Recipient{ Name : "masterCost", Channel : make(chan []byte) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- newCostResponseRecipient;
	
	for {
		select {
			case orderEncoded := <- newOrderRecipient.Channel:

				handleEventNewOrder(orderEncoded);
			
			case cost := <- newCostResponseRecipient.Channel:

				handleEventCostResponse();

			case 

					broadcastChannel <- network.Message{ RecipientName : "receiveNewDestinationOrder", Data : orderEncoded };				
		}	
	}
}
