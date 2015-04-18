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

		if !ordersUnconfirmed.AlreadyStored(order) && !ordersGlobal.AlreadyStored(order) {
			
			ordersUnconfirmed.Add(order);
			orderEncoded, _ := JSON.Encode(order);

			transmitChannel <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);
		}
	}
}

func workerHandleNewDestinationOrder(transmitChannel chan network.Message, message network.Message, elevatorEventNewOrder chan Order) {
	
	log.Data("Worker: Got new desitination order");

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {}

	if ordersUnconfirmed.AlreadyStored(order) {
		ordersUnconfirmed.Remove(order);
	}

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(order);
	}

	elevatorEventNewOrder <- order;

	transmitChannel <- network.MakeMessage("distributorOrderTakenConfirmation", message.Data, distributorIPAddr);
}

func workerHandleDestinationOrderTakenBySomeone(message network.Message) {

	log.Data("Worker: Some has taken a order");

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	if ordersUnconfirmed.AlreadyStored(order) {
		ordersUnconfirmed.Remove(order);
	}

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(order);
	}
}

//-----------------------------------------------//

func workerHandleCostRequest(message network.Message, elevatorEventCostRequest chan Order) {
	
	log.Data("Worker: Got request for cost of order");

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	elevatorEventCostRequest <- order;
}

func workerHandleElevatorCostResponse(cost int, transmitChannel chan network.Message) {

	log.Data("Worker: Cost from local is", cost);

	costEncoded, _ := JSON.Encode(cost);
	transmitChannel <- network.MakeMessage("distributorCostResponse", costEncoded, distributorIPAddr);
}

//-----------------------------------------------//

func workerHandleDistributorChange(message network.Message) {

	log.Data("Worker: I have a new distributor now", message.SenderIPAddr);
	distributorIPAddr = message.SenderIPAddr;
}