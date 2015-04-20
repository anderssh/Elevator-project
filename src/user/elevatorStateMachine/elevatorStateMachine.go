package elevatorStateMachine

import(
	. "user/typeDefinitions"
	"user/config"
	"user/elevator"
	"user/log"
	"user/ordersLocal"
	"time"
	"fmt"
);

//-----------------------------------------------//

type State int

const (
	STATE_STARTUP   	State = iota
	STATE_IDLE      	State = iota
	STATE_MOVING    	State = iota
	STATE_DOOR_OPEN 	State = iota
);

//-----------------------------------------------//

var currentState 		State;
var floorDestination 	int;
var displaySwitch		bool;

var eventReachedNewFloor 		chan int;
var eventCloseDoor 				chan bool;
var eventStop 					chan bool;
var eventObstruction 			chan bool;
var eventButtonFloorPressed 	chan ButtonFloor;

var eventNewDestinationOrder 	chan Order;

var eventOrdersExecutedOnFloorBySomeone chan int;
var eventDestinationOrderTakenBySomeone chan Order;
var eventRemoveCallUpAndCallDownOrders chan bool;

var workerNewOrder 				chan Order;
var workerOrdersExecutedOnFloor chan int;

var eventCostRequest			chan Order;
var workerCostResponse 			chan int;

var workerExitsStartup 			chan bool;

//-----------------------------------------------//

func Initialize(eventNewDestinationOrderArg chan Order,
				eventCostRequestArg chan Order,
				eventOrdersExecutedOnFloorBySomeoneArg chan int,
				eventDestinationOrderTakenBySomeoneArg chan Order,
				eventRemoveCallUpAndCallDownOrdersArg chan bool,

				workerExitsStartupArg chan bool,
				workerNewOrderArg chan Order,
				workerCostResponseArg chan int,
				workerOrdersExecutedOnFloorArg chan int) {

	err := elevator.Initialize();

	if err != nil {
		log.Error(err);
	}

	eventReachedNewFloor 	= make(chan int);
	eventCloseDoor 			= make(chan bool);
	eventStop 				= make(chan bool);
	eventObstruction 		= make(chan bool);
	eventButtonFloorPressed = make(chan ButtonFloor);
	
	eventNewDestinationOrder 	= eventNewDestinationOrderArg;

	eventCostRequest 			= eventCostRequestArg;
	eventOrdersExecutedOnFloorBySomeone = eventOrdersExecutedOnFloorBySomeoneArg;
	eventDestinationOrderTakenBySomeone = eventDestinationOrderTakenBySomeoneArg;
	eventRemoveCallUpAndCallDownOrders = eventRemoveCallUpAndCallDownOrdersArg;

	workerExitsStartup 			= workerExitsStartupArg;
	workerNewOrder 				= workerNewOrderArg;
	workerCostResponse 			= workerCostResponseArg;
	workerOrdersExecutedOnFloor = workerOrdersExecutedOnFloorArg;

	currentState 	 = STATE_STARTUP;
	floorDestination = -1;

	elevator.DriveInDirection(DIRECTION_DOWN);
}

//-----------------------------------------------//

func Display() {
	
	if config.SHOULD_DISPLAY_ELEVATOR {

		//-----------------------------------------------//
		// Title with state

		fmt.Print("ELEVATOR: ");

		switch currentState {
			case STATE_STARTUP:
				fmt.Println(" (STATE_STARTUP) ");
			case STATE_IDLE:
				fmt.Println(" (STATE_IDLE) ");
			case STATE_MOVING:
				fmt.Println(" (STATE_MOVING) ");
			case STATE_DOOR_OPEN:
				fmt.Println(" (STATE_DOOR_OPEN) ");
		}

		//-----------------------------------------------//

		for floor := 3; floor >= 0; floor-- {

			//-----------------------------------------------//
			// Elevator position

			fmt.Print(floor);
			fmt.Print("|"); // Left wall

			if elevator.GetLastReachedFloor() == floor {
				fmt.Print("\x1b[33;1m");
				fmt.Print("+++");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("   ");
			}

			fmt.Print("|"); // Right wall

			if currentState == STATE_DOOR_OPEN && elevator.GetLastReachedFloor() == floor {
				fmt.Print("->");
			} else {
				fmt.Print("  ");
			}

			//-----------------------------------------------//
			// Order display

			fmt.Print("\t");

			if ordersLocal.AlreadyStored(Order{ Type : ORDER_INSIDE, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("O");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("O");
			}

			fmt.Print(" ");

			if ordersLocal.AlreadyStored(Order{ Type : ORDER_UP, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("^");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("^");
			}

			fmt.Print(" ");

			if ordersLocal.AlreadyStored(Order{ Type : ORDER_DOWN, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("_");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("_");
			}

			fmt.Print("\n");
		}
	}
}

//-----------------------------------------------//

func handleEventReachedNewFloor(floorReached int) {
	
	switch currentState {
		case STATE_STARTUP:

			elevator.Stop()
			elevator.SetLastReachedFloor(floorReached);

			workerExitsStartup <- true;

			currentState = STATE_IDLE;

		case STATE_IDLE:

			elevator.SetLastReachedFloor(floorReached);

		case STATE_MOVING:
			
			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();

				if floorDestination == floorReached {
				
					elevator.Stop();

					elevator.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true
					});

					currentState = STATE_DOOR_OPEN;

				} else { // Update if other elevators has taken the orders

					if floorDestination > elevator.GetLastReachedFloor() {
						elevator.DriveInDirection(DIRECTION_UP);
					} else {
						elevator.DriveInDirection(DIRECTION_DOWN);
					}
				}

			} else {

				elevator.Stop();
				currentState = STATE_IDLE;
			}

			elevator.SetLastReachedFloor(floorReached);

		case STATE_DOOR_OPEN:

			elevator.SetLastReachedFloor(floorReached);
	}
}

//-----------------------------------------------//

func handleEventCloseDoor() {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Closed door in state startup");

		case STATE_IDLE:

			log.Warning("Closed door in state idle");

		case STATE_MOVING:

			log.Warning("Closed door in state moving");

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveOnFloor(elevator.GetLastReachedFloor());
			workerOrdersExecutedOnFloor <- elevator.GetLastReachedFloor();
			
			elevator.TurnOffLightDoorOpen();
			elevator.TurnOffAllLightButtonsOnFloor(elevator.GetLastReachedFloor());

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();
			
				if floorDestination == elevator.GetLastReachedFloor() {
					
					log.Warning("Orders still on floor when door closed.");
					
					elevator.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true
					});

					currentState = STATE_DOOR_OPEN;

				} else {

					if floorDestination > elevator.GetLastReachedFloor() {
						elevator.DriveInDirection(DIRECTION_UP);
					} else {
						elevator.DriveInDirection(DIRECTION_DOWN);
					}

					currentState = STATE_MOVING;
				}
			} else {
				currentState = STATE_IDLE;
			}
	}
}

//-----------------------------------------------//

func handleEventButtonPressed(button ButtonFloor) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to order at startup. Order not registered.");

		case STATE_IDLE:

			workerNewOrder <- button.ConvertToOrder();

		case STATE_MOVING:

			workerNewOrder <- button.ConvertToOrder();

		case STATE_DOOR_OPEN:

			workerNewOrder <- button.ConvertToOrder();
	}
}

//-----------------------------------------------//

func handleEventNewDestinationOrder(order Order) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle order at startup. Order not registered.");

		case STATE_IDLE:

			if !ordersLocal.AlreadyStored(order) {

				ordersLocal.Add(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());
				elevator.TurnOnLightButtonFromOrder(order);
			}

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();
				
				if (floorDestination == elevator.GetLastReachedFloor()) {
					
					elevator.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true;
					});

					currentState = STATE_DOOR_OPEN;

				} else if floorDestination < elevator.GetLastReachedFloor() {
					
					elevator.DriveInDirection(DIRECTION_DOWN);
					currentState = STATE_MOVING;

				} else {

					elevator.DriveInDirection(DIRECTION_UP);
					currentState = STATE_MOVING;
				}
			}

		case STATE_MOVING:

			if !ordersLocal.AlreadyStored(order) {
				
				ordersLocal.Add(order, elevator.GetLastReachedFloor(), true, elevator.GetDirection());
				elevator.TurnOnLightButtonFromOrder(order);

				floorDestination = ordersLocal.GetDestination();
			}

		case STATE_DOOR_OPEN:

			if !ordersLocal.AlreadyStored(order) {
				
				ordersLocal.Add(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());
				elevator.TurnOnLightButtonFromOrder(order);

				floorDestination = ordersLocal.GetDestination();
			}
	}
}

//-----------------------------------------------//

func handleDestinationOrderTakenBySomeone(order Order) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle order taken by someone at startup.");

		case STATE_IDLE:

			elevator.TurnOnLightButtonFromOrder(order);

		case STATE_MOVING:

			elevator.TurnOnLightButtonFromOrder(order);

		case STATE_DOOR_OPEN:

			elevator.TurnOnLightButtonFromOrder(order);
	}
}

//-----------------------------------------------//

func handleEventOrdersExectuedOnFloorBySomeone(floor int) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle event orders executed by someone at startup.");

		case STATE_IDLE:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevator.TurnOffCallUpAndCallDownLightButtonsOnFloor(floor);

		case STATE_MOVING:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevator.TurnOffCallUpAndCallDownLightButtonsOnFloor(floor);

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevator.TurnOffCallUpAndCallDownLightButtonsOnFloor(floor);
	}
}

//-----------------------------------------------//

func handleEventCostRequest(order Order) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to get cost when in startup. Nothing happens");

		case STATE_IDLE:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());

		case STATE_MOVING:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), true, elevator.GetDirection());

		case STATE_DOOR_OPEN:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());
	}
}

//-----------------------------------------------//

func handleEventRemoveCallUpAndCallDownOrders() {

		switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried remove call up and down orders at startup.");

		case STATE_IDLE:

			ordersLocal.RemoveCallUpAndCallDown();

		case STATE_MOVING:

			ordersLocal.RemoveCallUpAndCallDown();

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveCallUpAndCallDown();
	}
}

//-----------------------------------------------//

func stateMachine() {

	for {
		select {
			case <- eventStop:
				
				log.Warning("Pushed stop button. Button has no effect.");

			case <- eventObstruction:

				log.Warning("Pushed obstruction button. Button has no effect.");

			case floorReached := <- eventReachedNewFloor:

				handleEventReachedNewFloor(floorReached);
				Display();

			case <- eventCloseDoor:

				handleEventCloseDoor();
				Display();

			case button := <- eventButtonFloorPressed:

				handleEventButtonPressed(button);
				Display();

			case order := <- eventNewDestinationOrder:

				handleEventNewDestinationOrder(order);
				Display();

			case order := <- eventDestinationOrderTakenBySomeone:

				handleDestinationOrderTakenBySomeone(order);
				Display();

			case floor := <- eventOrdersExecutedOnFloorBySomeone:

				handleEventOrdersExectuedOnFloorBySomeone(floor);
				Display();

			case <- eventRemoveCallUpAndCallDownOrders:

				handleEventRemoveCallUpAndCallDownOrders();

			case order := <- eventCostRequest:

				handleEventCostRequest(order);
				Display();
		}
	}
}

//-----------------------------------------------//

func Run() {

	go elevator.RegisterEvents(	eventReachedNewFloor,
								eventStop,
								eventObstruction,
								eventButtonFloorPressed);

	go stateMachine();
}