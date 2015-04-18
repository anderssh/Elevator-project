package elevatorController;

import(
	. "user/typeDefinitions"
	"user/network"
	"user/config"
	"user/log"
	"time"
	"user/encoder/JSON"
	"strings"
	"strconv"
);

//-----------------------------------------------//

type State int

const (
	STATE_IDLE   								State = iota
	STATE_AWAITING_COST_RESPONSE   				State = iota
	STATE_AWAITING_ORDER_TAKEN_CONFIRMATION		State = iota

	STATE_AWAITING_MERGE_DATA					State = iota

	STATE_INACTIVE 								State = iota
);

var currentState State;

//-----------------------------------------------//

var workerIPAddrs []string;

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
// Order handling

var currentlyHandledOrder Order = Order{ -1, -1 };

func distributorHandleNewOrder(message network.Message, transmitChannel chan network.Message) {
	
	switch currentState {
		case STATE_IDLE:
			
			log.Data("Distributor: Got new order to distribute")

			orderEncoded := message.Data;
			err := JSON.Decode(orderEncoded, &currentlyHandledOrder);

			if err != nil {
				log.Error(err, "decode error");
			}

			currentState = STATE_AWAITING_COST_RESPONSE;

			for worker := range workerIPAddrs {
				transmitChannel <- network.MakeMessage("workerCostRequest", orderEncoded, workerIPAddrs[worker]);
			}

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

	}
}

func distributorHandleCostResponse(message network.Message, transmitChannel chan network.Message){

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
				transmitChannel <- network.MakeMessage("workerNewDestinationOrder", order, costBids[0].SenderIPAddr);

				currentState = STATE_AWAITING_ORDER_TAKEN_CONFIRMATION;
			}

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

	}
}

//-----------------------------------------------//

func distributorHandleOrderTakenConfirmation(message network.Message, transmitChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:

		case STATE_AWAITING_COST_RESPONSE:

		case STATE_AWAITING_ORDER_TAKEN_CONFIRMATION:

			log.Data("Distributor: Got order taken Confirmation")

			var takenOrder Order;
			err := JSON.Decode(message.Data, &takenOrder);

			if err != nil{
				log.Error(err, "Decode error");
			}

			// Distribute to others for global storage


			// Clean up
			costBids = make([]CostBid, 0, 1);
			currentlyHandledOrder = Order{ -1, -1 };

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//
// Merging

func distributorHandleActiveNotificationTick(broadcastChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:

			message, _ := JSON.Encode("Alive");
			broadcastChannel <- network.MakeMessage("distributorActiveNotification", message, network.BROADCAST_ADDR);
	}
}

func distributorHandleDistributorDisconnect(timeoutDistributorActiveNotification *time.Timer, eventInactiveDisconnect chan string, eventChangeNotificationRecipientID chan string) {

	switch currentState {
		case STATE_INACTIVE:

			log.Data("No distributor, switch");

			timeoutDistributorActiveNotification.Stop();
			
			workerIPAddrs = make([]string, 0, 1);
			workerIPAddrs = append(workerIPAddrs, network.GetLocalIPAddr());

			eventChangeNotificationRecipientID <- "distributorActiveNotification";

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//

func distributorHandleActiveNotification(message network.Message, timeoutDistributorActiveNotification *time.Timer, transmitChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:
			
			IPAddrNumbersLocal := strings.Split(network.GetLocalIPAddr(), ".");
			IPAddrNumbersSender := strings.Split(message.SenderIPAddr, ".");
			
			IPAddrEndingLocal, _ := strconv.Atoi(IPAddrNumbersLocal[3]);
			IPAddrEndingSender, _ := strconv.Atoi(IPAddrNumbersSender[3]);

			if IPAddrEndingLocal > IPAddrEndingSender {

				log.Data("Distributor: Merge with", message.SenderIPAddr);
				messageMerge, _ := JSON.Encode("Merge");
				transmitChannel <- network.MakeMessage("distributorMergeRequest", messageMerge, message.SenderIPAddr);

				currentState = STATE_AWAITING_MERGE_DATA;
			}

		case STATE_INACTIVE:

			timeoutDistributorActiveNotification.Reset(config.MASTER_ALIVE_NOTIFICATION_TIMEOUT);
	}
}

func distributorHandleMergeRequest(message network.Message, transmitChannel chan network.Message) {

	switch currentState {
		case STATE_IDLE:

			log.Data("Distributor: Going into inactive some else is my distributor now.");

			workerIPAddrsEncoded, _ := JSON.Encode(workerIPAddrs);

			transmitChannel <- network.MakeMessage("distributorMergeData", workerIPAddrsEncoded, message.SenderIPAddr);

			currentState = STATE_INACTIVE;
	}
}

func distributorHandleMergeData(message network.Message, transmitChannel chan network.Message) {

	switch currentState {

		case STATE_AWAITING_MERGE_DATA:

			log.Data("Distributor: Merge data received");

			var newWorkerIPAddrs []string;
			err := JSON.Decode(message.Data, &newWorkerIPAddrs);

			if err != nil {
				log.Error(err);
			}

			workerIPAddrs = append(workerIPAddrs, newWorkerIPAddrs ...);

			for worker := range newWorkerIPAddrs {
				log.Data("Distributor: has new worker", newWorkerIPAddrs[worker]);
				transmitChannel <- network.MakeMessage("workerChangeDistributor", make([]byte, 0, 1), newWorkerIPAddrs[worker]);
			}

			currentState = STATE_IDLE;
	}
}

//-----------------------------------------------//

func distributorHandleInactiveNotification(message network.Message) {

	/*_, keyExists := inactiveDisconnectTimeouts[message.SenderIPAddr];

	if keyExists {
			
		inactiveDisconnectTimeouts[message.SenderIPAddr].Reset(config.SLAVE_ALIVE_NOTIFICATION_TIMEOUT);
		log.Data("Inactive in list, reset...");
	}*/
}

func distributorHandleInactiveDisconnect(workerDisconnectIP string) {

	// Redistribute order of disconnected node
	log.Data("Disconnected worker", workerDisconnectIP);
}