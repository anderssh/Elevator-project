package elevatorController;

import(
	. "../typeDefinitions"
	"../network"
	"../encoder/JSON"
	"../log"
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

			broadcastChannel <- network.Message{ RecipientName : "masterNewOrder", Data : orderEncoded };
		}
	}
}

func handleNewDestinationOrder(orderEncoded []byte, elevatorEventNewOrder chan Order) {
	
	var order Order;
	err := JSON.Decode(orderEncoded, &order);

	if err != nil {}

	removeUnconfirmedOrder(order);
	elevatorEventNewOrder <- order;
}

func handleMasterCostRequest(order Order) {
	
	// Get cost
	// Return on master network
}

//-----------------------------------------------//
 
func slave(broadcastChannel 		  chan network.Message,
		   addServerRecipientChannel  chan network.Recipient,
		   elevatorOrderReceiver 	  chan Order,
		   elevatorEventNewOrder      chan Order) {

	newMasterDestinationOrderRecipient := network.Recipient{ Name : "receiveNewDestinationOrder", Channel : make(chan []byte) };
	masterCostRequestRecipient  := network.Recipient{ Name : "requestCostOfOrder", 		Channel : make(chan []byte) };

	addServerRecipientChannel <- newMasterDestinationOrderRecipient;
	addServerRecipientChannel <- masterCostRequestRecipient;
	
	for {
		select {
			case order := <- elevatorOrderReceiver:
				handleNewOrder(order, broadcastChannel, elevatorEventNewOrder);
			case orderEncoded := <- newMasterDestinationOrderRecipient.Channel:
				handleMasterNewDestinationOrder(orderEncoded, elevatorEventNewOrder);
			case orderEncoded := <- masterCostRequestRecipient.Channel:
				handleMasterCostRequest(orderEncoded);

		}
	}
}