package systemController;

import(
	. "../typeDefinitions"
	"../elevatorController"
);

//-----------------------------------------------//

var orderHandler 	chan Order = make(chan Order);
var eventNewOrder 	chan Order = make(chan Order);

//-----------------------------------------------//

func Run() {

	elevatorController.Initialize(orderHandler, eventNewOrder);
	elevatorController.Run();

	for {
		select {
			case order := <- orderHandler:
				eventNewOrder <- order;
		}
	}
}