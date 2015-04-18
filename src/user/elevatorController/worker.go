package elevatorController;

import(
	. "user/typeDefinitions"
	"user/network"
	"user/encoder/JSON"
	"user/log"
	"user/ordersGlobal"
	"user/ordersUnconfirmed"
);

//-----------------------------------------------//

var distributorIPAddr string = network.GetLocalIPAddr();

//-----------------------------------------------//

func workerHandleEventNewOrder(order Order, transmitChannel chan network.Message, elevatorEventNewOrder chan Order) {
	
	if order.Type == ORDER_INSIDE { 							// Should only be dealt with locally
		
		elevatorEventNewOrder <- order;
		
	} else {

		if !ordersUnconfirmed.AlreadyStored(order) {
			
			ordersUnconfirmed.Add(order);
			orderEncoded, _ := JSON.Encode(order);

			transmitChannel <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);
		}
	}
}

func workerHandleEventNewDestinationOrder(message network.Message, elevatorEventNewOrder chan Order) {
	
	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {}

	ordersUnconfirmed.Remove(order);

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(order);
	}

	elevatorEventNewOrder <- order;
}

func workerHandleCostRequest(message network.Message, elevatorEventCostRequest chan Order) {
	
	log.Data("Worker: Got request for cost of order")

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	elevatorEventCostRequest <- order;
}

func workerHandleElevatorCostResponse(cost int, transmitChannel chan network.Message) {

	costEncoded, _ := JSON.Encode(cost);
	log.Data("Worker: Cost from local is", cost);
	transmitChannel <- network.MakeMessage("distributorCostResponse", costEncoded, distributorIPAddr);
}