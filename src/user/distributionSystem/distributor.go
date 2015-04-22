package distributionSystem;

import(
	. "user/typeDefinitions"
	"user/network"
	"user/config"
	"user/log"
	"user/orders/ordersGlobal"
	"user/encoder/JSON"
	"strings"
	"strconv"
	"time"
);

//-----------------------------------------------//

type State int

const (
	STATE_STARTUP 								State = iota
	STATE_IDLE   								State = iota
	
	STATE_AWAITING_COST_RESPONSE   				State = iota
	STATE_AWAITING_ORDER_TAKEN_CONFIRMATION		State = iota

	STATE_AWAITING_MERGE_DATA					State = iota

	STATE_INACTIVE 								State = iota
);

var currentState State;

var currentlyHandledOrder OrderLocal = OrderLocal{ -1, -1 };

//-----------------------------------------------//

var workerIPAddrs []string;

func removeIpAddrFromWorkerIpAddrList(remoteAddr string) {	

	for worker := range workerIPAddrs {
		if (workerIPAddrs[worker] == remoteAddr) {
			workerIPAddrs = append(workerIPAddrs[:worker], workerIPAddrs[worker + 1:] ...);
			return;
		}
	}
}

func distributorDisplayWorkers() {

	if config.SHOULD_DISPLAY_WORKERS {

		log.DataWithColor(log.COLOR_CYAN, "-------------------------");
		log.DataWithColor(log.COLOR_CYAN, "Workers:");

		for worker := range workerIPAddrs {

			if workerIPAddrs[worker] == network.GetLocalIPAddr() {
				log.Data(workerIPAddrs[worker], "| Local");
			} else {
				log.Data(workerIPAddrs[worker]);
			}
		}

		log.DataWithColor(log.COLOR_CYAN, "-------------------------");
	}
}

//-----------------------------------------------//

type CostBid struct {
	Value			int
	SenderIPAddr 	string
}

var costBids []CostBid;

func costBidAllreadyStored(costBid CostBid) bool {
		
	for costBidIndex := 0 ; costBidIndex < len(costBids); costBidIndex++{
		if (costBids[costBidIndex].SenderIPAddr == costBid.SenderIPAddr) {
			return true;
		}
	}
	
	return false;
}

func costBidAddAndSort(newCostBid CostBid) {
	
	costBids = append(costBids, newCostBid);
		
	for costBidIndex := (len(costBids) - 1); costBidIndex > 0; costBidIndex--{
		
		tempCostBid := costBids[costBidIndex]

		if (costBids[costBidIndex].Value < costBids[costBidIndex-1].Value) {
			
			costBids[costBidIndex] 		= costBids[costBidIndex-1];
			costBids[costBidIndex-1] 	= tempCostBid;
		}
	}
}

//-----------------------------------------------//

func distributorInitialize(transmitChannelTCP chan network.Message) {

	localIPAddr := network.GetLocalIPAddr();

	workerIPAddrs = make([]string, 0, 1);
	workerIPAddrs = append(workerIPAddrs, localIPAddr);

	costBids = make([]CostBid, 0, 1);

	currentlyHandledOrder = OrderLocal{ -1, -1 };

	ordersGlobalEncoded, _ := JSON.Encode(ordersGlobal.GetAll());

	transmitChannelTCP <- network.MakeMessage("workerChangeDistributor", ordersGlobalEncoded, localIPAddr);
}

//-----------------------------------------------//

func returnToStateIdle(eventRedistributeOrder chan bool) {

	if ordersGlobal.HasOrderToRedistribute() {
		time.AfterFunc(time.Millisecond * 50, func() { eventRedistributeOrder <- true });
	}

	currentState = STATE_IDLE;
}

//-----------------------------------------------//

func distributorHandleElevatorExitsStartup(eventRedistributeOrder chan bool) {

	log.Data("Distributor: Exits startup");
	
	returnToStateIdle(eventRedistributeOrder);
}

//-----------------------------------------------//
// Order handling

func distributorHandleRedistributionOfOrder(transmitChannelTCP chan network.Message) {
		
	switch currentState {
		case STATE_IDLE:

			if ordersGlobal.HasOrderToRedistribute() {

				log.Data("Distributor: Got new order to redistribute.")

				orderToRedistribute := ordersGlobal.GetOrderToRedistribute();
			
				currentlyHandledOrder = OrderLocal{ Type : orderToRedistribute.Type, Floor : orderToRedistribute.Floor };
				orderEncoded, _ := JSON.Encode(currentlyHandledOrder);

				currentState = STATE_AWAITING_COST_RESPONSE;

				for worker := range workerIPAddrs {
					transmitChannelTCP <- network.MakeMessage("workerCostRequest", orderEncoded, workerIPAddrs[worker]);
				}
			}

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

	}
}

func distributorHandleNewOrder(message network.Message, transmitChannelTCP chan network.Message) {
	
	switch currentState {
		case STATE_IDLE:
			
			log.Error("Take")
			log.Data("Distributor: Got new order to distribute")

			orderEncoded := message.Data;
			err := JSON.Decode(orderEncoded, &currentlyHandledOrder);

			if err != nil {
				log.Error(err, "decode error");
			}

			currentState = STATE_AWAITING_COST_RESPONSE;

			for worker := range workerIPAddrs {
				transmitChannelTCP <- network.MakeMessage("workerCostRequest", orderEncoded, workerIPAddrs[worker]);
			}

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

	}
}

func distributorHandleCostResponse(message network.Message, transmitChannelTCP chan network.Message){

	switch currentState {
		case STATE_IDLE:
			
		case STATE_AWAITING_COST_RESPONSE:

			var cost int;
			err := JSON.Decode(message.Data, &cost);

			if err != nil {
				log.Error(err);
			}

			log.Data("Distributor: Got cost", cost, "from", message.SenderIPAddr);

			newCostBid := CostBid{ Value : cost, SenderIPAddr : message.SenderIPAddr };

			if !costBidAllreadyStored(newCostBid) {
				costBidAddAndSort(newCostBid);
			}

			if len(workerIPAddrs) == len(costBids) {

				log.Data("Distributor: send destination", currentlyHandledOrder.Floor, "to", costBids[0].SenderIPAddr);
				
				order, _ := JSON.Encode(currentlyHandledOrder);
				transmitChannelTCP <- network.MakeMessage("workerNewDestinationOrder", order, costBids[0].SenderIPAddr);

				currentState = STATE_AWAITING_ORDER_TAKEN_CONFIRMATION;
			}

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

	}
}

//-----------------------------------------------//

func distributorHandleOrderTakenConfirmation(message network.Message, transmitChannelTCP chan network.Message, eventRedistributeOrder chan bool) {

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

			log.Data("Distributor: Got order taken Confirmation");

			var order OrderLocal;
			err := JSON.Decode(message.Data, &order);

			if err != nil {
				log.Error(err, "Decode error");
			}

			// Distribute to others for global storage
			orderGlobal := ordersGlobal.MakeFromOrderLocal(order, message.SenderIPAddr);
			orderGlobalEncoded, _ := JSON.Encode(orderGlobal);

			for costBidIndex := 1; costBidIndex < len(costBids); costBidIndex++ {
				transmitChannelTCP <- network.MakeMessage("workerDestinationOrderTakenBySomeone", orderGlobalEncoded, costBids[costBidIndex].SenderIPAddr);
			}

			// Clean up
			costBids = make([]CostBid, 0, 1);
			currentlyHandledOrder = OrderLocal{ -1, -1 };

			returnToStateIdle(eventRedistributeOrder);
	}
}

//-----------------------------------------------//

func distributorHandleOrdersExecutedOnFloor(message network.Message, transmitChannelTCP chan network.Message) {

	switch currentState {

		case STATE_STARTUP:

		default:
			
			log.Data("Distributor: orders on floor executed by someone");

			for worker := range workerIPAddrs {
				transmitChannelTCP <- network.MakeMessage("workerOrdersExecutedOnFloorBySomeone", message.Data, workerIPAddrs[worker]);
			}
	}
}

//-----------------------------------------------//
// Merging

func distributorHandleNotificationTick(transmitChannelUDP chan network.Message) {

	switch currentState {
		case STATE_IDLE:
			transmitChannelUDP <- network.MakeMessage("distributorActiveNotification", make([]byte, 0, 1), network.BROADCAST_ADDR);
	}
}

//-----------------------------------------------//

type MergeData struct {
	WorkerIPAddrs 	[]string
	Orders 			[]OrderGlobal
}

var mergeIPAddr string;

func distributorHandleActiveNotification(message network.Message, transmitChannelTCP chan network.Message) {

	switch currentState {
		case STATE_IDLE:
			
			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				currentState = STATE_AWAITING_MERGE_DATA;

				mergeIPAddr = message.SenderIPAddr;

				log.Data("Distributor: Merge with", mergeIPAddr);
				transmitChannelTCP <- network.MakeMessage("distributorMergeRequest", make([]byte, 0, 1), mergeIPAddr);
			}
	}
}

func distributorHandleMergeRequest(message network.Message, transmitChannelTCP chan network.Message) {

	switch currentState {
		case STATE_IDLE:

			log.Data("Distributor: Going into inactive some else is my distributor now.", message.SenderIPAddr);

			mergeData := MergeData{ WorkerIPAddrs : workerIPAddrs, Orders : ordersGlobal.GetAll() };
			mergeDataEncoded, _ := JSON.Encode(mergeData);

			transmitChannelTCP <- network.MakeMessage("distributorMergeData", mergeDataEncoded, message.SenderIPAddr);

			distributorIPAddr = message.SenderIPAddr;

			currentState = STATE_INACTIVE;
	}
}

func distributorHandleMergeData(message network.Message, transmitChannelTCP chan network.Message, eventRedistributeOrder chan bool) {

	switch currentState {

		case STATE_AWAITING_MERGE_DATA:

			log.Data("Distributor: Merge data received");

			var mergeData MergeData;
			err := JSON.Decode(message.Data, &mergeData);

			if err != nil {
				log.Error(err);
			}

			ordersGlobal.MergeWith(mergeData.Orders);
			ordersGlobal.ResetAllResponsibilities();

			log.Data("Distributor: merged, now notify new workers");
			workerIPAddrs = append(workerIPAddrs, mergeData.WorkerIPAddrs ...);

			ordersGlobalEncoded, _ := JSON.Encode(ordersGlobal.GetAll());

			for worker := range mergeData.WorkerIPAddrs {
				log.Data("Distributor: has new worker", mergeData.WorkerIPAddrs[worker]);
				transmitChannelTCP <- network.MakeMessage("workerChangeDistributor", ordersGlobalEncoded, mergeData.WorkerIPAddrs[worker]);
			}

			mergeIPAddr = "";

			returnToStateIdle(eventRedistributeOrder);
	}
}

//-----------------------------------------------//

func distributorHandleElevatorNotFunctional(message network.Message, eventRedistributeOrder chan bool) {

	log.Data("Distirbutor: Elevator is not functioning properly, redistribute", message.SenderIPAddr);

	ordersGlobal.ResetResponsibilityOnWorkerIPAddr(message.SenderIPAddr);

	switch currentState {
		case STATE_IDLE:
			returnToStateIdle(eventRedistributeOrder);
	}
}

//-----------------------------------------------//
// Disconnect

func distributorHandleConnectionCheck(transmitChannelTCP chan network.Message) { 				// Force disconnect event if someone has disconnected

	switch currentState {

		case STATE_INACTIVE:

			transmitChannelTCP <- network.MakeMessage("check", make([]byte, 0, 1), distributorIPAddr);

		default:

			for worker := range workerIPAddrs {
				transmitChannelTCP <- network.MakeMessage("check", make([]byte, 0, 1), workerIPAddrs[worker]);
			}
	}
}

func distributorHandleConnectionDisconnect(disconnectIPAddr string, transmitChannelTCP chan network.Message, eventRedistributeOrder chan bool) {

	switch currentState {
		case STATE_IDLE:

			log.Data("Distributor: disconnected in idle", disconnectIPAddr);
			
			removeIpAddrFromWorkerIpAddrList(disconnectIPAddr);

			ordersGlobal.ResetResponsibilityOnWorkerIPAddr(disconnectIPAddr);

			returnToStateIdle(eventRedistributeOrder);

		case STATE_AWAITING_COST_RESPONSE:

			log.Data("Distributor: disconnected while waiting for cost response, aborting...");

			costBids = make([]CostBid, 0, 1);
			currentlyHandledOrder = OrderLocal{ -1, -1 };

			removeIpAddrFromWorkerIpAddrList(disconnectIPAddr);

			ordersGlobal.ResetResponsibilityOnWorkerIPAddr(disconnectIPAddr);

			returnToStateIdle(eventRedistributeOrder);

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

			log.Data("Distributor: disconnected while waiting for order taken confirmation, aborting...");

			costBids = make([]CostBid, 0, 1);
			currentlyHandledOrder = OrderLocal{ -1, -1 };

			removeIpAddrFromWorkerIpAddrList(disconnectIPAddr);

			ordersGlobal.ResetResponsibilityOnWorkerIPAddr(disconnectIPAddr);

			returnToStateIdle(eventRedistributeOrder);

		case STATE_AWAITING_MERGE_DATA:

			if disconnectIPAddr == mergeIPAddr {

				mergeIPAddr = "";

				returnToStateIdle(eventRedistributeOrder);
			}

		case STATE_INACTIVE:

			if disconnectIPAddr == distributorIPAddr {

				log.Data("Distributor: disconnected in INACTIVE. I am now a distributor.");

				distributorInitialize(transmitChannelTCP);

				ordersGlobal.ResetAllResponsibilities();

				returnToStateIdle(eventRedistributeOrder);
			}
	}
}