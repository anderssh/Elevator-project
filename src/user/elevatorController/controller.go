package elevatorController;

import(
	. "user/typeDefinitions"
	"user/elevatorStateMachine"
	"user/network"
	"time"
	"user/config"
);

//-----------------------------------------------//

var workerOrderFromElevatorReceiver chan Order;
var workerCostResponseFromElevatorReceiver chan int;

var elevatorEventNewOrder chan Order;
var elevatorEventCostRequest chan Order;

//-----------------------------------------------//

func Initialize() {
	
	workerOrderFromElevatorReceiver = make(chan Order);
	elevatorEventNewOrder = make(chan Order);

	elevatorEventCostRequest = make(chan Order, 10);
	workerCostResponseFromElevatorReceiver = make(chan int, 10);

	elevatorStateMachine.Initialize(workerOrderFromElevatorReceiver,
									elevatorEventNewOrder,
									elevatorEventCostRequest,
									workerCostResponseFromElevatorReceiver);
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
	go network.TCPTransmitServer(transmitChannel);

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

	addServerRecipientChannel <- distributorNewOrderRecipient;
	addServerRecipientChannel <- distributorCostResponseRecipient;
	addServerRecipientChannel <- distributorOrderTakenConfirmationRecipient;

	distributorMergeRequestRecipient 			:= network.Recipient{ ID : "distributorMergeRequest", 			ReceiveChannel : make(chan network.Message) };
	distributorMergeDataRecipient 				:= network.Recipient{ ID : "distributorMergeData", 				ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- distributorMergeRequestRecipient;
	addServerRecipientChannel <- distributorMergeDataRecipient;

	//-----------------------------------------------//

	distributorActiveNotificationRecipient := network.Recipient{ ID : "distributorActiveNotification", 		ReceiveChannel : make(chan network.Message) };

	addBroadcastRecipientChannel <- distributorActiveNotificationRecipient;

	eventDistributorActiveNotificationTicker := time.NewTicker(config.MASTER_ALIVE_NOTIFICATION_DELAY);

	//-----------------------------------------------//

	eventInactiveDisconnect 						:= make(chan string);
	eventDsitributorActiveNotificationTimeout 		:= make(chan bool);
	eventChangeNotificationRecipientID 				:= make(chan string);

	timeoutDistributorActiveNotification := time.AfterFunc(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT, func() {
		//eventDsitributorActiveNotificationTimeout <- true;
	});

	//-----------------------------------------------//
	// Worker setup

	workerNewDestinationOrderRecipient 	:= network.Recipient{ ID : "workerNewDestinationOrder", ReceiveChannel : make(chan network.Message) };
	workerCostRequestRecipient 		   	:= network.Recipient{ ID : "workerCostRequest", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- workerNewDestinationOrderRecipient;
	addServerRecipientChannel <- workerCostRequestRecipient;

	workerChangeDistributor 		   	:= network.Recipient{ ID : "workerChangeDistributor", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- workerChangeDistributor;

	//-----------------------------------------------//

	for {
		select {

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Distributor
			//-----------------------------------------------//
			//-----------------------------------------------//

			//-----------------------------------------------//
			// Distribute order

			case message := <- distributorNewOrderRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleNewOrder(message, transmitChannel);
			
			case message := <- distributorCostResponseRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleCostResponse(message, transmitChannel);

			case message := <- distributorOrderTakenConfirmationRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleOrderTakenConfirmation(message, transmitChannel);

			//-----------------------------------------------//
			// Distributor switching

			case <- eventDistributorActiveNotificationTicker.C:

				distributorHandleActiveNotificationTick(broadcastChannel);

			case  <- eventDsitributorActiveNotificationTimeout:

				distributorHandleDistributorDisconnect(timeoutDistributorActiveNotification, eventInactiveDisconnect, eventChangeNotificationRecipientID);

			//-----------------------------------------------//

			case message := <- distributorActiveNotificationRecipient.ReceiveChannel:
				
				distributorHandleActiveNotification(message, timeoutDistributorActiveNotification, transmitChannel);

			case message := <- distributorMergeRequestRecipient.ReceiveChannel:

				distributorHandleMergeRequest(message, transmitChannel);

			case message := <- distributorMergeDataRecipient.ReceiveChannel:

				distributorHandleMergeData(message, transmitChannel);

			//-----------------------------------------------//
			// Inactive registration

			case inactiveDisconnectIP := <- eventInactiveDisconnect:

				distributorHandleInactiveDisconnect(inactiveDisconnectIP);

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Worker
			//-----------------------------------------------//
			//-----------------------------------------------//
			
			//-----------------------------------------------//
			// Orders

			case order := <- workerOrderFromElevatorReceiver:
				
				workerHandleEventNewOrder(order, transmitChannel, elevatorEventNewOrder);
			
			case message := <- workerCostRequestRecipient.ReceiveChannel:
				
				workerHandleCostRequest(message, elevatorEventCostRequest);

			case cost := <- workerCostResponseFromElevatorReceiver:

				workerHandleElevatorCostResponse(cost, transmitChannel);

			case message := <- workerNewDestinationOrderRecipient.ReceiveChannel:
				
				workerHandleNewDestinationOrder(transmitChannel, message, elevatorEventNewOrder);
			
			//-----------------------------------------------//

			case message := <- workerChangeDistributor.ReceiveChannel:

				workerHandleDistributorChange(message);
		}
	}

}