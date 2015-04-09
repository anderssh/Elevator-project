package elevatorController;

import(
	. "../typeDefinitions"
	"../elevatorStateMachine"
	"../network"
);

//-----------------------------------------------//

var elevatorOrderReceiver chan Order;
var elevatorEventNewOrder chan Order;

//-----------------------------------------------//

func Initialize() {
	
	elevatorOrderReceiver = make(chan Order);
	elevatorEventNewOrder = make(chan Order);

	elevatorStateMachine.Initialize(elevatorOrderReceiver, elevatorEventNewOrder);
}

func Run() {

	elevatorStateMachine.Run();

	addServerRecipientChannel 	:= make(chan network.Recipient);
	broadcastChannel 			:= make(chan network.Message);

	go network.ListenServer("", 9123, addServerRecipientChannel);
	go network.TransmitServer("255.255.255.255", 9123, broadcastChannel);

	go master(broadcastChannel, addServerRecipientChannel);
	go slave(broadcastChannel, addServerRecipientChannel, elevatorOrderReceiver, elevatorEventNewOrder);
}