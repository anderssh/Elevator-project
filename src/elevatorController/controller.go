package elevatorController;

import(
	. "../typeDefinitions"
	"../elevatorStateMachine"
	"../network"
);

//-----------------------------------------------//

var elevatorOrderReceiver chan Order;
var elevatorEventNewOrder chan Order;

var elevatorEventCostRequest chan Order;
var elevatorCostResponseReceiver chan int;

//-----------------------------------------------//

func Initialize() {
	
	elevatorOrderReceiver = make(chan Order);
	elevatorEventNewOrder = make(chan Order);

	elevatorEventCostRequest = make(chan Order);
	elevatorCostResponseReceiver = make(chan int);

	elevatorStateMachine.Initialize(elevatorOrderReceiver,
									elevatorEventNewOrder,
									elevatorEventCostRequest,
									elevatorCostResponseReceiver);
}

func Run() {

	elevatorStateMachine.Run();

	addServerRecipientChannel 	:= make(chan network.Recipient);
	broadcastChannel 			:= make(chan network.Message);

	go network.ListenServer("", addServerRecipientChannel);
	go network.TransmitServer(broadcastChannel);

	go master(broadcastChannel, addServerRecipientChannel);
	go slave(broadcastChannel, addServerRecipientChannel, elevatorOrderReceiver, elevatorEventNewOrder, elevatorEventCostRequest, elevatorCostResponseReceiver);
}