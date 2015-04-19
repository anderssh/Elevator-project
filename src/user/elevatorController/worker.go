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

func workerHandleElevatorNewOrder(order Order, transmitChannel chan network.Message, elevatorEventNewDestinationOrder chan Order, eventUnconfirmedOrderTimeout chan Order) {
	
	if order.Type == ORDER_INSIDE { 							// Should only be dealt with locally
		
		elevatorEventNewDestinationOrder <- order;
		
	} else {

		if !ordersUnconfirmed.AlreadyStored(order) && !ordersGlobal.AlreadyStored(order) {
			
			ordersUnconfirmed.Add(order, eventUnconfirmedOrderTimeout);
			orderEncoded, _ := JSON.Encode(order);

			transmitChannel <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);
		}
	}

	ordersGlobal.Display();
}

//-----------------------------------------------//

func workerHandleEventUnconfirmedOrderTimeout(order Order, transmitChannel chan network.Message, elevatorEventNewDestinationOrder chan Order, eventUnconfirmedOrderTimeout chan Order) {

	log.Data("Worker: Did not receive confirmation on the order I sent up");

	orderEncoded, _ := JSON.Encode(order);
	ordersUnconfirmed.ResetTimer(order, eventUnconfirmedOrderTimeout)

	transmitChannel <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);

}

//-----------------------------------------------//

func workerHandleNewDestinationOrder(transmitChannel chan network.Message, message network.Message, elevatorEventNewDestinationOrder chan Order) {
	
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

	elevatorEventNewDestinationOrder <- order;

	transmitChannel <- network.MakeMessage("distributorOrderTakenConfirmation", message.Data, distributorIPAddr);
}

func workerHandleDestinationOrderTakenBySomeone(message network.Message, elevatorDestinationOrderTakenBySomeone chan Order) {

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

	elevatorDestinationOrderTakenBySomeone <- order;
}

//-----------------------------------------------//

func workerHandleElevatorOrdersExecutedOnFloor(floor int, transmitChannel chan network.Message) {

	log.Data("Worker: Executed orders on floor", floor);

	ordersGlobal.RemoveOnFloor(floor);

	floorEncoded, _ := JSON.Encode(floor);

	transmitChannel <- network.MakeMessage("distributorOrdersExecutedOnFloor", floorEncoded, distributorIPAddr);
}

func workerHandleOrdersExecutedOnFloorBySomeone(message network.Message, elevatorOrdersExecutedOnFloorBySomeone chan int) {

	log.Data("Worker: Someone has handled order on floor");

	var floor int;
	err := JSON.Decode(message.Data, &floor);

	if err != nil {
		log.Error(err);
	}

	ordersGlobal.RemoveOnFloor(floor);
	
	elevatorOrdersExecutedOnFloorBySomeone <- floor;

	ordersGlobal.Display();
}

//-----------------------------------------------//

func workerHandleCostRequest(message network.Message, elevatorCostRequest chan Order) {
	
	log.Data("Worker: Got request for cost of order");

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	elevatorCostRequest <- order;
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