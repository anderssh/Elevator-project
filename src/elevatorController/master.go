package elevatorController;

import(
	//"../typeDefinitions"
	"../network"
	//"../encoder/JSON"
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

var currentState State = STATE_IDLE;

//-----------------------------------------------//

func handleEventNewOrder(message network.Message, broadcastChannel chan network.Message) {
	
	orderEncoded := message.Data;

	switch currentState {
		case STATE_IDLE:

			broadcastChannel <- network.MakeMessage("slaveOrderRequest", orderEncoded, network.BROADCAST_ADDR);

			currentState = STATE_AWAITING_COST_RESPONSE;

		case STATE_AWAITING_COST_RESPONSE:
	}
}

func handleEventCostResponse(message network.Message){

}

func master(broadcastChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	orderRecipient := network.Recipient{ Name : "masterNewOrder", ReceiveChannel : make(chan network.Message) };
	costResponseRecipient := network.Recipient{ Name : "masterCost", ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- orderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	
	for {
		select {
			case message := <- orderRecipient.ReceiveChannel:

				handleEventNewOrder(message, broadcastChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				handleEventCostResponse(message);
		}	
	}
}
