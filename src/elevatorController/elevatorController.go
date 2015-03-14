package elevatorController

import (
	. "../typeDefinitions"
	"../elevator"
	"../log"
	"../orders"
	"time"
)

//-----------------------------------------------//

type State int

const (
	STATE_STARTUP   State = iota
	STATE_IDLE      State = iota
	STATE_MOVING    State = iota
	STATE_DOOR_OPEN State = iota
)

//-----------------------------------------------//

var currentState 		State;
var floorDestination 	int;
var floorLastVisited 	int;

var eventReachedNewFloor 	chan int 			= make(chan int)
var eventCloseDoor 			chan bool 			= make(chan bool)
var eventStop 				chan bool 			= make(chan bool)
var eventObstruction 		chan bool 			= make(chan bool)
var eventButtonFloorPressed chan ButtonFloor 	= make(chan ButtonFloor)
var eventNewOrder 			chan Order;

var orderHandler 			chan Order;

//-----------------------------------------------//

func handleEventReachedNewFloor(floorReached int) {
	
	switch currentState {
		case STATE_STARTUP:

			elevator.Stop()

			floorLastVisited 	= floorReached;
			currentState 		= STATE_IDLE;

		case STATE_IDLE:

			floorLastVisited 	= floorReached;

		case STATE_MOVING:

			if floorDestination == floorReached {

				elevator.Stop()
				currentState = STATE_DOOR_OPEN

				time.AfterFunc(time.Second*3, func() { // Close the door
					eventCloseDoor <- true
				});
			}

			floorLastVisited = floorReached;

		case STATE_DOOR_OPEN:

			floorLastVisited = floorReached;
	}
}

//-----------------------------------------------//

func handleEventCloseDoor() {
	
	switch currentState {
		case STATE_STARTUP:

			// Nothing

		case STATE_IDLE:

			// Nothing

		case STATE_MOVING:

			// Nothing

		case STATE_DOOR_OPEN:

			orders.RemoveTop();

			if orders.Exists() {

				floorDestination = orders.GetDestination();
			
				if floorDestination == floorLastVisited {
					currentState = STATE_IDLE;
				} else {

					if floorDestination > floorLastVisited {
						elevator.DriveInDirection(elevator.DIRECTION_UP);
					} else {
						elevator.DriveInDirection(elevator.DIRECTION_DOWN);
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

			// Nothing

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

			// Nothing

		case STATE_IDLE:

			if !orders.AllreadyStored(order) {
				orders.Add(order);
			}

			if orders.Exists() {

				floorDestination = orders.GetDestination();
				
				if (floorDestination == floorLastVisited) {
					
					currentState = STATE_DOOR_OPEN;
					time.AfterFunc(time.Second*3, func() { // Close the door
						eventCloseDoor <- true
					});

				} else if floorDestination < floorLastVisited {
					elevator.DriveInDirection(elevator.DIRECTION_DOWN);
				} else {
					elevator.DriveInDirection(elevator.DIRECTION_UP);
				}

				currentState = STATE_MOVING;
			}

		case STATE_MOVING:

			if !orders.AllreadyStored(order) {
				orders.Add(order);
			}

		case STATE_DOOR_OPEN:

			if !orders.AllreadyStored(order) {
				orders.Add(order);
			}
	}
}

//-----------------------------------------------//

func stateMachine() {

	for {
		select {
			case floorReached := <- eventReachedNewFloor:
				handleEventReachedNewFloor(floorReached);
			case <- eventCloseDoor:
				handleEventCloseDoor();
			case <- eventStop:
				// Not handled
			case <- eventObstruction:
				// Not handled
			case button := <- eventButtonFloorPressed:
				handleEventButtonPressed(button);
			case order := <- eventNewOrder:
				handleEventNewOrder(order);
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

	elevator.DriveInDirection(elevator.DIRECTION_DOWN);
}

func Run() {

	go elevator.RegisterEvents(	eventReachedNewFloor,
								eventStop,
								eventObstruction,
								eventButtonFloorPressed);

	go stateMachine();
}
