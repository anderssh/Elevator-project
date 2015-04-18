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

	newOrderRecipient 				:= network.Recipient{ ID : "distributorNewOrder", 				ReceiveChannel : make(chan network.Message) };
	costResponseRecipient 			:= network.Recipient{ ID : "distributorCostResponse", 			ReceiveChannel : make(chan network.Message) };
	orderTakenConfirmationRecipient := network.Recipient{ ID : "distributorOrderTakenConfirmation", 	ReceiveChannel : make(chan network.Message) };

	mergeRequestRecipient 			:= network.Recipient{ ID : "distributorMergeRequest", 			ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newOrderRecipient;
	addServerRecipientChannel <- costResponseRecipient;
	addServerRecipientChannel <- orderTakenConfirmationRecipient;

	addServerRecipientChannel <- mergeRequestRecipient;

	//-----------------------------------------------//

	activeNotificationRecipient := network.Recipient{ ID : "distributorActiveNotification", 		ReceiveChannel : make(chan network.Message) };

	addBroadcastRecipientChannel <- activeNotificationRecipient;

	eventActiveNotificationTicker := time.NewTicker(config.MASTER_ALIVE_NOTIFICATION_DELAY);

	//-----------------------------------------------//

	eventInactiveDisconnect 			:= make(chan string);
	eventActiveNotificationTimeout 		:= make(chan bool);
	eventChangeNotificationRecipientID 	:= make(chan string);

	timeoutDistributorActiveNotification := time.AfterFunc(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT, func() {
		//eventActiveNotificationTimeout <- true;
	});

	//-----------------------------------------------//
	// Worker setup

	eventChangeDistributor 				:= make(chan string);

	newDestinationOrderRecipient := network.Recipient{ ID : "workerNewDestinationOrder", ReceiveChannel : make(chan network.Message) };
	costRequestRecipient 		 := network.Recipient{ ID : "workerCostRequest", 		ReceiveChannel : make(chan network.Message) };

	addServerRecipientChannel <- newDestinationOrderRecipient;
	addServerRecipientChannel <- costRequestRecipient;

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

			case message := <- newOrderRecipient.ReceiveChannel:

				distributorDisplayWorkers();
				distributorHandleEventNewOrder(message, transmitChannel);
			
			case message := <- costResponseRecipient.ReceiveChannel:

				distributorHandleEventCostResponse(message, transmitChannel);

			case message := <- orderTakenConfirmationRecipient.ReceiveChannel:

				distributorHandleEventOrderTakenConfirmation(message, transmitChannel);

			//-----------------------------------------------//
			// Distributor switching

			case <- eventActiveNotificationTicker.C:

				distributorHandleActiveNotificationTick(broadcastChannel);

			case message := <- activeNotificationRecipient.ReceiveChannel:
				
				distributorHandleActiveNotification(message, timeoutDistributorActiveNotification, transmitChannel);

			case message := <- mergeRequestRecipient.ReceiveChannel:

				distributorHandleMergeRequest(message, eventChangeDistributor);

			case  <- eventActiveNotificationTimeout:

				distributorHandleDistributorDisconnect(timeoutDistributorActiveNotification, eventInactiveDisconnect, eventChangeNotificationRecipientID);

			//-----------------------------------------------//
			// Inactive registration

			case inactiveDisconnectIP := <- eventInactiveDisconnect:

				distributorHandleInactiveDisconnect(inactiveDisconnectIP);

			//-----------------------------------------------//
			//-----------------------------------------------//
			// Worker
			//-----------------------------------------------//
			//-----------------------------------------------//
			
			case order := <- elevatorOrderReceiver:
				
				workerHandleEventNewOrder(order, transmitChannel, elevatorEventNewOrder);
			
			case message := <- newDestinationOrderRecipient.ReceiveChannel:
				
				workerHandleEventNewDestinationOrder(message, elevatorEventNewOrder);
			
			case message := <- costRequestRecipient.ReceiveChannel:
				
				workerHandleCostRequest(message, elevatorEventCostRequest);

			case cost := <- elevatorCostResponseReceiver:

				workerHandleElevatorCostResponse(cost, transmitChannel);

			case newDistributorIPAddr := <- eventChangeDistributor:

				log.Data("Worker: I have a new distributor now", newDistributorIPAddr);
				distributorIPAddr = newDistributorIPAddr;
		}
	}

}