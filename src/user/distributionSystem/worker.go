package distributionSystem;

import(
	. "user/typeDefinitions"
	"user/network"
	"user/encoder/JSON"
	"user/log"
	"user/orders/ordersGlobal"
	"user/orders/ordersUnconfirmed"
);

//-----------------------------------------------//

var distributorIPAddr string = network.GetLocalIPAddr();

//-----------------------------------------------//

func workerHandleElevatorNewOrder(order OrderLocal, transmitChannelTCP chan network.Message, elevatorEventNewDestinationOrder chan OrderLocal, eventUnconfirmedOrderTimeout chan OrderLocal) {
	
	if order.Type == ORDER_INSIDE { 							// Should only be dealt with locally
		
		elevatorEventNewDestinationOrder <- order;
		
	} else {

		if !ordersUnconfirmed.AlreadyStored(order) && !ordersGlobal.AlreadyStored(order) {
			
			ordersUnconfirmed.Add(order, eventUnconfirmedOrderTimeout);
			orderEncoded, _ := JSON.Encode(order);

			transmitChannelTCP <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);
		}
	}

	ordersGlobal.Display();
}

//-----------------------------------------------//

func workerHandleEventUnconfirmedOrderTimeout(order OrderLocal, transmitChannelTCP chan network.Message, elevatorEventNewDestinationOrder chan OrderLocal, eventUnconfirmedOrderTimeout chan OrderLocal) {

	//log.Data("Worker: Did not receive confirmation on the order I sent up");

	orderEncoded, _ := JSON.Encode(order);
	ordersUnconfirmed.ResetTimer(order, eventUnconfirmedOrderTimeout);

	transmitChannelTCP <- network.MakeMessage("distributorNewOrder", orderEncoded, distributorIPAddr);
}

//-----------------------------------------------//

func workerHandleNewDestinationOrder(transmitChannelTCP chan network.Message, message network.Message, elevatorEventNewDestinationOrder chan OrderLocal) {
	
	log.Data("Worker: Got new desitination order");

	var order OrderLocal;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	orderGlobal := ordersGlobal.MakeFromOrderLocal(order, network.GetLocalIPAddr());

	if ordersUnconfirmed.AlreadyStored(order) {
		ordersUnconfirmed.Remove(order);
	}

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(orderGlobal);
	} else {
		ordersGlobal.UpdateResponsibility(orderGlobal);
	}

	elevatorEventNewDestinationOrder <- order;

	transmitChannelTCP <- network.MakeMessage("distributorOrderTakenConfirmation", message.Data, distributorIPAddr);
}

func workerHandleDestinationOrderTakenBySomeone(message network.Message, elevatorDestinationOrderTakenBySomeone chan OrderLocal) {

	log.Data("Worker: Some has taken a order");

	var orderGlobal OrderGlobal;
	err := JSON.Decode(message.Data, &orderGlobal);

	if err != nil {
		log.Error(err);
	}

	order := OrderLocal{ Type : orderGlobal.Type, Floor : orderGlobal.Floor };

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(orderGlobal);
	} else {
		ordersGlobal.UpdateResponsibility(orderGlobal);
	}

	if ordersUnconfirmed.AlreadyStored(order) {
		ordersUnconfirmed.Remove(order);
	}

	elevatorDestinationOrderTakenBySomeone <- order;
}

//-----------------------------------------------//

func workerHandleElevatorOrdersExecutedOnFloor(floor int, transmitChannelTCP chan network.Message) {

	log.Data("Worker: Executed orders on floor", floor);

	ordersGlobal.RemoveOnFloor(floor);

	floorEncoded, _ := JSON.Encode(floor);

	transmitChannelTCP <- network.MakeMessage("distributorOrdersExecutedOnFloor", floorEncoded, distributorIPAddr);
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

func workerHandleCostRequest(message network.Message, elevatorCostRequest chan OrderLocal) {
	
	log.Data("Worker: Got request for cost of order");

	var order OrderLocal;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	elevatorCostRequest <- order;
}

func workerHandleElevatorCostResponse(cost int, transmitChannelTCP chan network.Message) {

	log.Data("Worker: Cost from local is", cost);

	costEncoded, _ := JSON.Encode(cost);
	transmitChannelTCP <- network.MakeMessage("distributorCostResponse", costEncoded, distributorIPAddr);
}

//-----------------------------------------------//

func workerHandleDistributorChange(message network.Message, elevatorRemoveCallUpAndCallDownOrders chan bool) {

	log.Data("Worker: I have a new distributor now", message.SenderIPAddr, "delete all call up and down orders.");
	
	distributorIPAddr = message.SenderIPAddr;

	log.Data("Worker: Updating global orderlist");

	var newOrdersGlobal []OrderGlobal;
	err := JSON.Decode(message.Data, &newOrdersGlobal);

	if err != nil {
		log.Error(err);
	}

	ordersGlobal.SetTo(newOrdersGlobal);

	elevatorRemoveCallUpAndCallDownOrders <- true;
}