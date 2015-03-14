package elevatorController

import(
	. "../typeDefinitions"
	"../elevator"
	"../log"
	"../orders"
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

var eventReachedNewFloor 	chan int 			= make(chan int)
var eventCloseDoor 			chan bool 			= make(chan bool)
var eventStop 				chan bool 			= make(chan bool)
var eventObstruction 		chan bool 			= make(chan bool)
var eventButtonFloorPressed chan ButtonFloor 	= make(chan ButtonFloor)
var eventNewOrder 			chan Order;

var orderHandler 			chan Order;

//-----------------------------------------------//

func Display() {
	
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

	for floor := 4; floor >= 1; floor-- {

		//-----------------------------------------------//
		// Elevator position

		fmt.Print(floor);
		fmt.Print("|"); // Left wall

		if elevator.GetPreviouslyReachedFloor() == floor {
			fmt.Print("\x1b[33;1m");
			fmt.Print("+++");
			fmt.Print("\x1b[0m");
		} else {
			fmt.Print("   ");
		}

		fmt.Print("|"); // Right wall

		if currentState == STATE_DOOR_OPEN && elevator.GetPreviouslyReachedFloor() == floor {
			fmt.Print("->");
		} else {
			fmt.Print("  ");
		}

		//-----------------------------------------------//
		// Order display

		fmt.Print("\t");

		if orders.AllreadyStored(Order{ Type : ORDER_INSIDE, Floor : floor }) {
			fmt.Print("\x1b[31;1m");
			fmt.Print("O");
			fmt.Print("\x1b[0m");
		} else {
			fmt.Print("O");
		}

		fmt.Print(" ");

		if orders.AllreadyStored(Order{ Type : ORDER_UP, Floor : floor }) {
			fmt.Print("\x1b[31;1m");
			fmt.Print("^");
			fmt.Print("\x1b[0m");
		} else {
			fmt.Print("^");
		}

		fmt.Print(" ");

		if orders.AllreadyStored(Order{ Type : ORDER_DOWN, Floor : floor }) {
			fmt.Print("\x1b[31;1m");
			fmt.Print("_");
			fmt.Print("\x1b[0m");
		} else {
			fmt.Print("_");
		}

		fmt.Print("\n");
	}
}

//-----------------------------------------------//

func handleEventReachedNewFloor(floorReached int) {
	
	switch currentState {
		case STATE_STARTUP:

			elevator.Stop()

			elevator.SetPreviouslyReachedFloor(floorReached);
			currentState 		= STATE_IDLE;

		case STATE_IDLE:

			elevator.SetPreviouslyReachedFloor(floorReached);

		case STATE_MOVING:

			if floorDestination == floorReached {

				elevator.Stop()
				currentState = STATE_DOOR_OPEN

				time.AfterFunc(time.Second*3, func() { // Close the door
					eventCloseDoor <- true
				});
			}

			elevator.SetPreviouslyReachedFloor(floorReached);

		case STATE_DOOR_OPEN:

			elevator.SetPreviouslyReachedFloor(floorReached);
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

			orders.RemoveOnFloor(floorLastVisited);

			if orders.Exists() {

				floorDestination = orders.GetDestination();
			
				if floorDestination == elevator.GetPreviouslyReachedFloor() {
					
					log.Warning("Orders still on floor when door closed.");
					currentState = STATE_DOOR_OPEN;
					time.AfterFunc(time.Second*3, func() { // Close the door
						eventCloseDoor <- true
					});

				} else {

					if floorDestination > elevator.GetPreviouslyReachedFloor() {
						elevator.DriveInDirection(elevator.DIRECTION_UP);
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

			orderHandler <- orders.OrderFromButtonFloor(button);

		case STATE_MOVING:

			orderHandler <- orders.OrderFromButtonFloor(button);

		case STATE_DOOR_OPEN:

			orderHandler <- orders.OrderFromButtonFloor(button);
	}
}

//-----------------------------------------------//

func handleEventNewOrder(order Order) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle order at startup. Order not registered.");

		case STATE_IDLE:

			if !orders.AllreadyStored(order) {
				orders.Add(order, floorLastVisited, currentState == STATE_MOVING, elevator.GetDirection());
			}

			if orders.Exists() {

				floorDestination = orders.GetDestination();
				
				if (floorDestination == elevator.GetPreviouslyReachedFloor()) {
					
					currentState = STATE_DOOR_OPEN;
					time.AfterFunc(time.Second*3, func() { // Close the door
						eventCloseDoor <- true;
					});

<<<<<<< HEAD
				} else if floorDestination < elevator.GetPreviouslyReachedFloor() {
					elevator.DriveInDirection(elevator.DIRECTION_DOWN);
=======
				} else if floorDestination < floorLastVisited {
					elevator.DriveInDirection(DIRECTION_DOWN);
>>>>>>> orderImprovement
					currentState = STATE_MOVING;
				} else {
					elevator.DriveInDirection(DIRECTION_UP);
					currentState = STATE_MOVING;
				}
			}

		case STATE_MOVING:

			if !orders.AllreadyStored(order) {
				orders.Add(order, floorLastVisited, currentState == STATE_MOVING, elevator.GetDirection());
				floorDestination = orders.GetDestination();
			}

		case STATE_DOOR_OPEN:

			if !orders.AllreadyStored(order) {
				orders.Add(order, floorLastVisited, currentState == STATE_MOVING, elevator.GetDirection());
				floorDestination = orders.GetDestination();
			}
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
		}
	}
}

//-----------------------------------------------//

func Initialize(orderHandlerArg chan Order, eventNewOrderArg chan Order) {

	err := elevator.Initialize();

	if err != nil {
		log.Error(err);
	}

	orderHandler 	= orderHandlerArg;
	eventNewOrder 	= eventNewOrderArg;

	currentState 	= STATE_STARTUP;

	elevator.DriveInDirection(DIRECTION_DOWN);
}

func Run() {

	go elevator.RegisterEvents(	eventReachedNewFloor,
								eventStop,
								eventObstruction,
								eventButtonFloorPressed);

	go stateMachine();
}
