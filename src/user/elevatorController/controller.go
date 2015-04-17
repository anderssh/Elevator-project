package elevatorController;

import(
	. "user/typeDefinitions"
	"user/elevatorStateMachine"
	"user/network"
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

	elevatorEventCostRequest = make(chan Order, 10);
	elevatorCostResponseReceiver = make(chan int, 10);

	elevatorStateMachine.Initialize(elevatorOrderReceiver,
									elevatorEventNewOrder,
									elevatorEventCostRequest,
									elevatorCostResponseReceiver);
}

func Run() {

	elevatorStateMachine.Run();

	addServerRecipientChannel 		:= make(chan network.Recipient);
	addBroadcastRecipientChannel 	:= make(chan network.Recipient);
	
	transmitChannel 				:= make(chan network.Message);
	broadcastChannel 				:= make(chan network.Message);

	eventDisconnect 				:= make(chan string);
	eventChangeMaster 				:= make(chan string);

	go network.TCPListenServer("", addServerRecipientChannel, eventDisconnect);
	go network.TCPTransmitServer(transmitChannel);

	go network.UDPListenServer("", addBroadcastRecipientChannel);
	go network.UDPTransmitServer(broadcastChannel);

	go master(transmitChannel, addServerRecipientChannel, broadcastChannel, addBroadcastRecipientChannel, eventChangeMaster);
	go slave(transmitChannel, addServerRecipientChannel, elevatorOrderReceiver, elevatorEventNewOrder, elevatorEventCostRequest, elevatorCostResponseReceiver, eventChangeMaster);
}