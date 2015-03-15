package systemController;

import(
	. "../typeDefinitions"
	"../elevatorController"
);

//-----------------------------------------------//

var elevatorOrderReceiver 	chan Order = make(chan Order);
var elevatorEventNewOrder 	chan Order = make(chan Order);

//-----------------------------------------------//

func handleNewOrder(order Order) {
	
	if order.Type == ORDER_INSIDE { 			// Should only be dealt with locally

		elevatorEventNewOrder <- order;
	
	} else {

		//Send to master
		elevatorEventNewOrder <- order;

	}
}

//-----------------------------------------------//

func Run() {

	elevatorController.Initialize(elevatorOrderReceiver, elevatorEventNewOrder);
	elevatorController.Run();

	for {
		select {
			case order := <- elevatorOrderReceiver:
				handleNewOrder(order);
		}
	}
}