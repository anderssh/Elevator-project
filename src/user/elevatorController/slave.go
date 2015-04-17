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

var masterIPAddr string;

//-----------------------------------------------//

func slaveHandleEventNewOrder(order Order, transmitChannel chan network.Message, elevatorEventNewOrder chan Order) {
	
	if order.Type == ORDER_INSIDE { 							// Should only be dealt with locally
		
		elevatorEventNewOrder <- order;
		
	} else {

		if !ordersUnconfirmed.AlreadyStored(order) {
			
			ordersUnconfirmed.Add(order);
			orderEncoded, _ := JSON.Encode(order);

			transmitChannel <- network.MakeMessage("masterNewOrder", orderEncoded, masterIPAddr);
		}
	}
}

func slaveHandleEventNewDestinationOrder(message network.Message, elevatorEventNewOrder chan Order) {
	
	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {}

	ordersUnconfirmed.Remove(order);

	if !ordersGlobal.AlreadyStored(order) {
		ordersGlobal.Add(order);
	}

	elevatorEventNewOrder <- order;
}

func slaveHandleCostRequest(message network.Message, elevatorEventCostRequest chan Order) {
	
	log.Data("Slave: Got request for cost of order")

	var order Order;
	err := JSON.Decode(message.Data, &order);

	if err != nil {
		log.Error(err);
	}

	elevatorEventCostRequest <- order;
}

func slaveHandleElevatorCostResponse(cost int, transmitChannel chan network.Message) {

	costEncoded, _ := JSON.Encode(cost);
	log.Data("Slave: Cost from local is", cost);
	transmitChannel <- network.MakeMessage("masterCostResponse", costEncoded, masterIPAddr);
}

//-----------------------------------------------//
 
func slave(transmitChannel 		  	  	chan network.Message,
		   addServerRecipientChannel  	chan network.Recipient,
		   elevatorOrderReceiver 	  	chan Order,
		   elevatorEventNewOrder      	chan Order,
		   elevatorEventCostRequest   	chan Order,
		   elevatorCostResponseReceiver	chan int,
		   eventChangeMaster 			chan string) {

	newDestinationOrderRecipient := network.Recipient{ ID : "slaveNewDestinationOrder", ReceiveChannel : make(chan network.Message) };
	costRequestRecipient 		 := network.Recipient{ ID : "slaveCostRequest", ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newDestinationOrderRecipient;
	addServerRecipientChannel <- costRequestRecipient;

	masterIPAddr = network.GetLocalIPAddr();
	
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

			case newMasterIPAddr := <- eventChangeMaster:

				log.Data("Slave: I have a new master now", newMasterIPAddr);
				masterIPAddr = newMasterIPAddr;
		}
	}
}