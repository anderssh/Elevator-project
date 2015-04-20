package elevatorController;

import(
	. "user/typeDefinitions"
	"user/elevatorStateMachine"
	"user/network"
	"time"
	"user/config"
);

//-----------------------------------------------//

var elevatorNewDestinationOrder chan Order;
var elevatorCostRequest chan Order;
var elevatorOrdersExecutedOnFloorBySomeone chan int;
var elevatorDestinationOrderTakenBySomeone chan Order;
var elevatorRemoveCallUpAndCallDownOrders chan bool;

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

	eventElevatorNewOrder 			= make(chan Order);
	eventElevatorCostResponse 		= make(chan int, 10);
	eventElevatorOrdersExecutedOnFloor = make(chan int);

	elevatorStateMachine.Initialize(elevatorNewDestinationOrder,
									elevatorCostRequest,
									elevatorOrdersExecutedOnFloorBySomeone,
									elevatorDestinationOrderTakenBySomeone,
									elevatorRemoveCallUpAndCallDownOrders,

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
	
	transmitChannel 				:= make(chan network.Message);
	broadcastChannel 				:= make(chan network.Message);

	eventDisconnect 				:= make(chan string);

	go network.TCPListenServer("", addServerRecipientChannel, eventDisconnect);
	go network.TCPTransmitServer(transmitChannel, eventDisconnect);

	go network.UDPListenServer("", addBroadcastRecipientChannel);
	go network.UDPTransmitServer(broadcastChannel);

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

	eventDistributorActiveNotificationTicker := time.NewTicker(config.MASTER_ALIVE_NOTIFICATION_DELAY);

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

	//-----------------------------------------------//

	for {
		select {

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Distributor
			//-----------------------------------------------//
			//-----------------------------------------------//

			case disconnectIPAddr := <- eventDisconnect:

				distributorHandleConnectionDisconnect(disconnectIPAddr, transmitChannel, eventRedistributeOrder);

			//-----------------------------------------------//
			// Distribute order

			case message := <- distributorNewOrderRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleNewOrder(message, transmitChannel);

			case <- eventRedistributeOrder:

				distributorHandleRedistributionOfOrder(transmitChannel);
			
			case message := <- distributorCostResponseRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleCostResponse(message, transmitChannel);

			case message := <- distributorOrderTakenConfirmationRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleOrderTakenConfirmation(message, transmitChannel, eventRedistributeOrder);

			//-----------------------------------------------//

			case message := <- distributorOrdersExecutedOnFloorRecipient.ReceiveChannel:

				distributorHandleOrdersExecutedOnFloor(message, transmitChannel);

			//-----------------------------------------------//
			// Distributor switching and merging

			case <- eventDistributorActiveNotificationTicker.C:

				distributorHandleActiveNotificationTick(broadcastChannel);

			case message := <- distributorActiveNotificationRecipient.ReceiveChannel:
				
				distributorHandleActiveNotification(message, transmitChannel);

			case message := <- distributorMergeRequestRecipient.ReceiveChannel:

				distributorHandleMergeRequest(message, transmitChannel);

			case message := <- distributorMergeDataRecipient.ReceiveChannel:

				distributorHandleMergeData(message, transmitChannel, eventRedistributeOrder);

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Worker
			//-----------------------------------------------//
			//-----------------------------------------------//
			
			//-----------------------------------------------//
			// Orders 

			case order := <- eventElevatorNewOrder:
				
				workerHandleElevatorNewOrder(order, transmitChannel, elevatorNewDestinationOrder);
			
			case message := <- workerCostRequestRecipient.ReceiveChannel:
				
				workerHandleCostRequest(message, elevatorCostRequest);

			case cost := <- eventElevatorCostResponse:

				workerHandleElevatorCostResponse(cost, transmitChannel);

			case message := <- workerNewDestinationOrderRecipient.ReceiveChannel:
				
				workerHandleNewDestinationOrder(transmitChannel, message, elevatorNewDestinationOrder);
			
			case message := <- workerDestinationOrderTakenBySomeoneRecipient.ReceiveChannel:

				workerHandleDestinationOrderTakenBySomeone(message, elevatorDestinationOrderTakenBySomeone);

			//-----------------------------------------------//

			case floor := <- eventElevatorOrdersExecutedOnFloor:

				workerHandleElevatorOrdersExecutedOnFloor(floor, transmitChannel);

			case message := <- workerOrdersExecutedOnFloorBySomeoneRecipient.ReceiveChannel:

				workerHandleOrdersExecutedOnFloorBySomeone(message, elevatorOrdersExecutedOnFloorBySomeone);

			//-----------------------------------------------//

			case message := <- workerChangeDistributorRecipient.ReceiveChannel:
				
				workerHandleDistributorChange(message, elevatorRemoveCallUpAndCallDownOrders);
		}
	}

}