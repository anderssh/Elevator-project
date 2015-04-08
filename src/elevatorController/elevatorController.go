package elevatorController;

import(
	. "../typeDefinitions"
	"../elevatorStateMachine"
	"../network"
	"../encoder/JSON"
);

//-----------------------------------------------//

var slaveOrderReceiver 		chan Order = make(chan Order);
var slaveOrderDeletion 		chan Order = make(chan Order);

var elevatorOrderReceiver 	chan Order = make(chan Order);
var elevatorEventNewOrder 	chan Order = make(chan Order);

//-----------------------------------------------//

var ordersUnconfirmed 		[]Order    = make([]Order, 1);

func ordersUnconfirmedExists(order Order) bool {

	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

//-----------------------------------------------//

func handleNewOrder(order Order, transmitChannel chan network.Message) {
	
	if order.Type == ORDER_INSIDE { 		// Should only be dealt with locally

		slaveOrderReceiver <- order;

	} else {

		if !ordersUnconfirmedExists(order) {
			
			ordersUnconfirmed = append(ordersUnconfirmed, order); 		// Store until it is handled by some master
			orderMessage, _ := JSON.Encode(order);

			transmitChannel <- network.Message{ Recipient : "master_new_order", Data : string(orderMessage) };
		}
	}
}

//-----------------------------------------------//

func slave(transmitChannel chan network.Message) {
	
	for {
		select {
			case order := <- elevatorOrderReceiver:
				handleNewOrder(order, transmitChannel);
			case order := <- slaveOrderReceiver:
				elevatorEventNewOrder <- order;
		}
	}
}

//-----------------------------------------------//

var isCurrentlyMaster bool = true;

func master(addServerRecipientChannel chan network.Recipient) {

	newOrderRecipient := network.Recipient{ Name : "master_new_order", Channel : make(chan string) };

	addServerRecipientChannel <- newOrderRecipient;
	
	for {

		if (isCurrentlyMaster) {

			// If other masters, merge
			// If not, do the same stuff


		} else {
			// Timeout of master, discover new
		}		
	}
}

//-----------------------------------------------//

func Run() {

	elevatorStateMachine.Initialize(elevatorOrderReceiver, elevatorEventNewOrder);
	elevatorStateMachine.Run();

	//-----------------------------------------------//

	addServerRecipientChannel := make(chan network.Recipient);

	go network.ListenServer("255.255.255.255", 10005, addServerRecipientChannel);

	//-----------------------------------------------//

	transmitChannel := make(chan network.Message);

	go network.TransmitServer("255.255.255.255", 10005, transmitChannel);

	//-----------------------------------------------//

	go master(addServerRecipientChannel);
	go slave(transmitChannel);
}