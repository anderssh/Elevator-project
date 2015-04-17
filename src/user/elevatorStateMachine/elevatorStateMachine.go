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

var eventReachedNewFloor 	chan int;
var eventCloseDoor 			chan bool;
var eventStop 				chan bool;
var eventObstruction 		chan bool;
var eventButtonFloorPressed chan ButtonFloor;

var eventNewOrder 			chan Order;
var orderHandler 			chan Order;

var eventCostRequest		chan Order;
var costResponseHandler 	chan int;

//-----------------------------------------------//

func Initialize(orderHandlerArg chan Order,
				eventNewOrderArg chan Order,
				eventCostRequestArg chan Order,
				costResponseHandlerArg chan int) {

	err := elevator.Initialize();

	if err != nil {
		log.Error(err);
	}

	displaySwitch = true;

	eventReachedNewFloor 	= make(chan int);
	eventCloseDoor 			= make(chan bool);
	eventStop 				= make(chan bool);
	eventObstruction 		= make(chan bool);
	eventButtonFloorPressed = make(chan ButtonFloor);
	
	orderHandler 	= orderHandlerArg;
	eventNewOrder 	= eventNewOrderArg;

	eventCostRequest = eventCostRequestArg;
	costResponseHandler = costResponseHandlerArg

	currentState 	= STATE_STARTUP;
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
			currentState 		= STATE_IDLE;

		case STATE_IDLE:

			elevator.SetLastReachedFloor(floorReached);

		case STATE_MOVING:

			if floorDestination == floorReached {

				elevator.Stop();
				currentState = STATE_DOOR_OPEN
				elevator.TurnOnLightDoorOpen();
				time.AfterFunc(time.Second*3, func() { // Close the door
					eventCloseDoor <- true
				});
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

			elevator.TurnOffLightDoorOpen();

			ordersLocal.RemoveOnFloor(elevator.GetLastReachedFloor());
			elevator.TurnOffAllLightButtonsOnFloor(elevator.GetLastReachedFloor());

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();
			
				if floorDestination == elevator.GetLastReachedFloor() {
					
					log.Warning("Orders still on floor when door closed.");
					currentState = STATE_DOOR_OPEN;
					time.AfterFunc(time.Second*3, func() { // Close the door
						eventCloseDoor <- true
					});

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

			orderHandler <- button.ConvertToOrder();

		case STATE_MOVING:

			orderHandler <- button.ConvertToOrder();

		case STATE_DOOR_OPEN:

			orderHandler <- button.ConvertToOrder();
	}
}

//-----------------------------------------------//

func handleEventNewOrder(order Order) {

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
					
					currentState = STATE_DOOR_OPEN;
					elevator.TurnOnLightDoorOpen();
					time.AfterFunc(time.Second*3, func() { // Close the door
						eventCloseDoor <- true;
					});

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

func handleEventCostRequest(order Order) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tride to get cost when in startup. Nothing happens");

		case STATE_IDLE:

			costResponseHandler <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());

		case STATE_MOVING:

			costResponseHandler <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), true, elevator.GetDirection());

		case STATE_DOOR_OPEN:

			costResponseHandler <- ordersLocal.GetCostOf(order, elevator.GetLastReachedFloor(), false, elevator.GetDirection());
	}
}

//-----------------------------------------------//

func stateMachine() {

	for {
		select {
			case floorReached := <- eventReachedNewFloor:

				handleEventReachedNewFloor(floorReached);
				Display();

			case <- eventCloseDoor:

				handleEventCloseDoor();
				Display();

			case <- eventStop:
				
				log.Warning("Pushed stop button. Button has no effect.");

			case <- eventObstruction:

				log.Warning("Pushed obstruction button. Button has no effect.");

			case button := <- eventButtonFloorPressed:

				handleEventButtonPressed(button);
				Display();

			case order := <- eventNewOrder:

				handleEventNewOrder(order);
				Display();

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