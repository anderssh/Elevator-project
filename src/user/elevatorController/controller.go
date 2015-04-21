package elevatorController;

import(
	. "user/typeDefinitions"
	"user/elevatorStateMachine"
	"user/network"
	"time"
	"user/config"
	"user/log"
);

//-----------------------------------------------//

var elevatorNewDestinationOrder chan Order;
var elevatorCostRequest chan Order;
var elevatorOrdersExecutedOnFloorBySomeone chan int;
var elevatorDestinationOrderTakenBySomeone chan Order;
var elevatorRemoveCallUpAndCallDownOrders chan bool;

var eventElevatorExitsStartup chan bool;
var eventElevatorNewOrder chan Order;
var eventElevatorCostResponse chan int;
var eventElevatorOrdersExecutedOnFloor chan int;

//-----------------------------------------------//

func Initialize() {
	
	elevatorNewDestinationOrder 	= make(chan Order);
	elevatorCostRequest 			= make(chan Order, 10);
	elevatorOrdersExecutedOnFloorBySomeone = make(chan int);
	elevatorDestinationOrderTakenBySomeone = make(chan Order);
	elevatorRemoveCallUpAndCallDownOrders = make(chan bool);

	eventElevatorExitsStartup 		= make(chan bool);
	eventElevatorNewOrder 			= make(chan Order);
	eventElevatorCostResponse 		= make(chan int, 10);
	eventElevatorOrdersExecutedOnFloor = make(chan int);

	elevatorStateMachine.Initialize(elevatorNewDestinationOrder,
									elevatorCostRequest,
									elevatorOrdersExecutedOnFloorBySomeone,
									elevatorDestinationOrderTakenBySomeone,
									elevatorRemoveCallUpAndCallDownOrders,

									eventElevatorExitsStartup,
									eventElevatorNewOrder,
									eventElevatorCostResponse,
									eventElevatorOrdersExecutedOnFloor);
}

func Run() {

	elevatorStateMachine.Run();

	//-----------------------------------------------//
	// Network setup

	addServerRecipientChannel 		:= make(chan network.Recipient);
	addBroadcastRecipientChannel 	:= make(chan network.Recipient);
	
	transmitChannelTCP 				:= make(chan network.Message);
	transmitChannelUDP 				:= make(chan network.Message);

	eventDisconnect 				:= make(chan string);

	go network.TCPListenServer("", addServerRecipientChannel, eventDisconnect);
	go network.TCPTransmitServer(transmitChannelTCP, eventDisconnect);

	go network.UDPListenServer("", addBroadcastRecipientChannel);
	go network.UDPTransmitServer(transmitChannelUDP);

	//-----------------------------------------------//
	// Distributor setup

	currentState = STATE_IDLE;

	workerIPAddrs = make([]string, 0, 1);
	workerIPAddrs = append(workerIPAddrs, network.GetLocalIPAddr());

	costBids = make([]CostBid, 0, 1);

	//-----------------------------------------------//

	distributorNewOrderRecipient 				:= network.Recipient{ ID : "distributorNewOrder", 				ReceiveChannel : make(chan network.Message) };
	distributorCostResponseRecipient 			:= network.Recipient{ ID : "distributorCostResponse", 			ReceiveChannel : make(chan network.Message) };
	distributorOrderTakenConfirmationRecipient  := network.Recipient{ ID : "distributorOrderTakenConfirmation", ReceiveChannel : make(chan network.Message) };
	distributorOrdersExecutedOnFloorRecipient  	:= network.Recipient{ ID : "distributorOrdersExecutedOnFloor", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- distributorNewOrderRecipient;
	addServerRecipientChannel <- distributorCostResponseRecipient;
	addServerRecipientChannel <- distributorOrderTakenConfirmationRecipient;
	addServerRecipientChannel <- distributorOrdersExecutedOnFloorRecipient;

	distributorMergeRequestRecipient 	:= network.Recipient{ ID : "distributorMergeRequest", 	ReceiveChannel : make(chan network.Message) };
	distributorMergeDataRecipient 		:= network.Recipient{ ID : "distributorMergeData", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- distributorMergeRequestRecipient;
	addServerRecipientChannel <- distributorMergeDataRecipient;

	eventRedistributeOrder := make(chan bool);

	//------------------------------	-----------------//

	distributorActiveNotificationRecipient := network.Recipient{ ID : "distributorActiveNotification", 		ReceiveChannel : make(chan network.Message) };

	addBroadcastRecipientChannel <- distributorActiveNotificationRecipient;

	eventDistributorActiveNotificationTicker := time.NewTicker(config.DISTRIBUTOR_ALIVE_NOTIFICATION_DELAY);

	//-----------------------------------------------//
	// Worker setup

	workerNewDestinationOrderRecipient 				:= network.Recipient{ ID : "workerNewDestinationOrder", 			ReceiveChannel : make(chan network.Message) };
	workerCostRequestRecipient 		   				:= network.Recipient{ ID : "workerCostRequest", 					ReceiveChannel : make(chan network.Message) };
	workerDestinationOrderTakenBySomeoneRecipient 	:= network.Recipient{ ID : "workerDestinationOrderTakenBySomeone", 	ReceiveChannel : make(chan network.Message) };
	workerOrdersExecutedOnFloorBySomeoneRecipient 	:= network.Recipient{ ID : "workerOrdersExecutedOnFloorBySomeone", 	ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- workerNewDestinationOrderRecipient;
	addServerRecipientChannel <- workerCostRequestRecipient;
	addServerRecipientChannel <- workerDestinationOrderTakenBySomeoneRecipient;
	addServerRecipientChannel <- workerOrdersExecutedOnFloorBySomeoneRecipient;

	workerChangeDistributorRecipient 	:= network.Recipient{ ID : "workerChangeDistributor", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- workerChangeDistributorRecipient;

	eventUnconfirmedOrderTimeout 		:= make(chan Order);

	//-----------------------------------------------//

	for {
		select {

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Distributor
			//-----------------------------------------------//
			//-----------------------------------------------//

			case <- eventElevatorExitsStartup:

				log.Data("Distributor: Exits startup");
				currentState = STATE_IDLE;

			case disconnectIPAddr := <- eventDisconnect:

				distributorHandleConnectionDisconnect(disconnectIPAddr, transmitChannelTCP, eventRedistributeOrder);

			//-----------------------------------------------//
			// Distribute order

			case message := <- distributorNewOrderRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleNewOrder(message, transmitChannelTCP);

			case <- eventRedistributeOrder:

				distributorHandleRedistributionOfOrder(transmitChannelTCP);
			
			case message := <- distributorCostResponseRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleCostResponse(message, transmitChannelTCP);

			case message := <- distributorOrderTakenConfirmationRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleOrderTakenConfirmation(message, transmitChannelTCP, eventRedistributeOrder);

			//-----------------------------------------------//

			case message := <- distributorOrdersExecutedOnFloorRecipient.ReceiveChannel:

				distributorHandleOrdersExecutedOnFloor(message, transmitChannelTCP);

			//-----------------------------------------------//
			// Distributor switching and merging

			case <- eventDistributorActiveNotificationTicker.C:

				distributorHandleActiveNotificationTick(transmitChannelUDP);

			case message := <- distributorActiveNotificationRecipient.ReceiveChannel:
				
				distributorHandleActiveNotification(message, transmitChannelTCP);

			case message := <- distributorMergeRequestRecipient.ReceiveChannel:

				distributorHandleMergeRequest(message, transmitChannelTCP);

			case message := <- distributorMergeDataRecipient.ReceiveChannel:

				distributorHandleMergeData(message, transmitChannelTCP, eventRedistributeOrder);

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Worker
			//-----------------------------------------------//
			//-----------------------------------------------//
			
			//-----------------------------------------------//
			// Orders 

			case order := <- eventElevatorNewOrder:
				
				workerHandleElevatorNewOrder(order, transmitChannelTCP, elevatorNewDestinationOrder, eventUnconfirmedOrderTimeout);

			case order := <- eventUnconfirmedOrderTimeout:

				workerHandleEventUnconfirmedOrderTimeout(order, transmitChannelTCP, elevatorNewDestinationOrder, eventUnconfirmedOrderTimeout);
			
			case message := <- workerCostRequestRecipient.ReceiveChannel:
				
				workerHandleCostRequest(message, elevatorCostRequest);

			case cost := <- eventElevatorCostResponse:

				workerHandleElevatorCostResponse(cost, transmitChannelTCP);

			case message := <- workerNewDestinationOrderRecipient.ReceiveChannel:
				
				workerHandleNewDestinationOrder(transmitChannelTCP, message, elevatorNewDestinationOrder);
			
			case message := <- workerDestinationOrderTakenBySomeoneRecipient.ReceiveChannel:

				workerHandleDestinationOrderTakenBySomeone(message, elevatorDestinationOrderTakenBySomeone);

			//-----------------------------------------------//

			case floor := <- eventElevatorOrdersExecutedOnFloor:

				workerHandleElevatorOrdersExecutedOnFloor(floor, transmitChannelTCP);

			case message := <- workerOrdersExecutedOnFloorBySomeoneRecipient.ReceiveChannel:

				workerHandleOrdersExecutedOnFloorBySomeone(message, elevatorOrdersExecutedOnFloorBySomeone);

			//-----------------------------------------------//

			case message := <- workerChangeDistributorRecipient.ReceiveChannel:
				
				workerHandleDistributorChange(message, elevatorRemoveCallUpAndCallDownOrders);
		}
	}

}