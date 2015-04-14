package elevatorController;

import(
	. "../typeDefinitions"
	"../network"
	"../encoder/JSON"
	//"../log"
);

//-----------------------------------------------//

var ordersUnconfirmed []Order = make([]Order, 0, 1);

func ordersUnconfirmedExists(order Order) bool {

	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

func removeUnconfirmedOrder(order Order) {

	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			if (orderIndex == len(ordersUnconfirmed) - 1) {
				ordersUnconfirmed = ordersUnconfirmed[:(len(ordersUnconfirmed) - 1)];
			} else {
				ordersUnconfirmed = append(ordersUnconfirmed[0:orderIndex], ordersUnconfirmed[orderIndex + 1:] ... );
			}
		}
	}
}

//-----------------------------------------------//

func handleNewOrder(order Order, broadcastChannel chan network.Message, elevatorEventNewOrder chan Order) {
	
	if order.Type == ORDER_INSIDE { 								// Should only be dealt with locally
		
		elevatorEventNewOrder <- order;
		
	} else {

		if !ordersUnconfirmedExists(order) {
			
			ordersUnconfirmed = append(ordersUnconfirmed, order); 	// Store until it is handled by some master
			orderEncoded, _ := JSON.Encode(order);

			message := network.MakeMessage("masterNewOrder", orderEncoded, "255.255.255.255")

			broadcastChannel <- message;
		}
	}
}

func handleMasterNewDestinationOrder(message network.Message, elevatorEventNewOrder chan Order) {
	
	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {}

	removeUnconfirmedOrder(order);
	elevatorEventNewOrder <- order;
}

func handleMasterCostRequest(message network.Message) {
	
	// Get cost
	// Return on master network
}

//-----------------------------------------------//
 
func slave(broadcastChannel 		  chan network.Message,
		   addServerRecipientChannel  chan network.Recipient,
		   elevatorOrderReceiver 	  chan Order,
		   elevatorEventNewOrder      chan Order) {

	newMasterDestinationOrderRecipient 	:= network.Recipient{ ID : "receiveNewDestinationOrder", ReceiveChannel : make(chan network.Message) };
	masterCostRequestRecipient  		:= network.Recipient{ ID : "requestCostOfOrder", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newMasterDestinationOrderRecipient;
	addServerRecipientChannel <- masterCostRequestRecipient;
	
	for {
		select {
			case order := <- elevatorOrderReceiver:
				handleNewOrder(order, broadcastChannel, elevatorEventNewOrder);
			case message := <- newMasterDestinationOrderRecipient.ReceiveChannel:
				handleMasterNewDestinationOrder(message, elevatorEventNewOrder);
			case message := <- masterCostRequestRecipient.ReceiveChannel:
				handleMasterCostRequest(message);

		}
	}
}