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

func slaveHandleEventNewOrder(order Order, transmitChannel chan network.Message, elevatorEventNewOrder chan Order) {
	
	if order.Type == ORDER_INSIDE { 								// Should only be dealt with locally
		
		elevatorEventNewOrder <- order;
		
	} else {

		if !ordersUnconfirmedExists(order) {
			
			ordersUnconfirmed = append(ordersUnconfirmed, order); 	// Store until it is handled by some master
			orderEncoded, _ := JSON.Encode(order);

			message := network.MakeMessage("masterNewOrder", orderEncoded, "255.255.255.255");

			transmitChannel <- message;
		}
	}
}

func slaveHandleEventNewDestinationOrder(message network.Message, elevatorEventNewOrder chan Order) {
	
	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {}

	removeUnconfirmedOrder(order);
	elevatorEventNewOrder <- order;
}

func slaveHandleCostRequest(message network.Message, elevatorEventCostRequest chan Order) {
	
	log.Data("Slave: Got request for cost of order")

	var order Order;
	err := JSON.Decode(message.Data, &order);

	log.Error(err);

	elevatorEventCostRequest <- order;
}

func slaveHandleElevatorCostResponse(cost int, transmitChannel chan network.Message) {

	costEncoded, _ := JSON.Encode(cost);
	log.Data("Slave: Cost from local", cost)
	transmitChannel <- network.MakeMessage("masterCostResponse", costEncoded, network.BROADCAST_ADDR);
}

//-----------------------------------------------//
 
func slave(transmitChannel 		  	  	chan network.Message,
		   addServerRecipientChannel  	chan network.Recipient,
		   elevatorOrderReceiver 	  	chan Order,
		   elevatorEventNewOrder      	chan Order,
		   elevatorEventCostRequest   	chan Order,
		   elevatorCostResponseReceiver	chan int) {

	newDestinationOrderRecipient := network.Recipient{ ID : "slaveNewDestinationOrder", ReceiveChannel : make(chan network.Message) };
	costRequestRecipient 		 := network.Recipient{ ID : "slaveCostRequest", ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newDestinationOrderRecipient;
	addServerRecipientChannel <- costRequestRecipient;
	
	for {
		select {
			case order := <- elevatorOrderReceiver:
				
				slaveHandleEventNewOrder(order, transmitChannel, elevatorEventNewOrder);
			
			case message := <- newDestinationOrderRecipient.ReceiveChannel:
				
				slaveHandleEventNewDestinationOrder(message, elevatorEventNewOrder);
			
			case message := <- costRequestRecipient.ReceiveChannel:
				
				slaveHandleCostRequest(message, elevatorEventCostRequest);

			case cost := <- elevatorCostResponseReceiver:

				slaveHandleElevatorCostResponse(cost, transmitChannel);

		}
	}
}