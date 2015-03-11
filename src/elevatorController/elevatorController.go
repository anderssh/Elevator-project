package elevatorController

import (
	"../elevator"
	"../log"
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

var currentState State
var destination int

var eventReachedNewFloor chan int = make(chan int)
var eventCloseDoor chan bool = make(chan bool)
var eventStop chan bool = make(chan bool)
var eventObstruction chan bool = make(chan bool)
var eventButtonFloorPressed chan elevator.ButtonFloor = make(chan elevator.ButtonFloor)

//-----------------------------------------------//

func handleEventReachedNewFloor(floorReached int) {
	switch currentState {
	case STATE_STARTUP:

		elevator.Stop()
		currentState = STATE_IDLE

	case STATE_IDLE:

		// Nothing

	case STATE_MOVING:

		if destination == floorReached {

			elevator.Stop()
			currentState = STATE_DOOR_OPEN

			time.AfterFunc(time.Second*3, func() { // Close the door
				eventCloseDoor <- true
			})
		}

	case STATE_DOOR_OPEN:

		// Nothing
	}
}

func handleEventCloseDoor() {
	switch currentState {
	case STATE_STARTUP:

		// Nothing

	case STATE_IDLE:

		// Nothing

	case STATE_MOVING:

		// Nothing

	case STATE_DOOR_OPEN:

	}
}

func handleEventButtonPressed(button elevator.ButtonFloor) {
	switch currentState {
	case STATE_STARTUP:

		// Nothing

	case STATE_IDLE:

		destination = button.Floor

	case STATE_MOVING:

		// Add to orders

	case STATE_DOOR_OPEN:

		// Add to orders if not on this floor
	}
}

//-----------------------------------------------//

func stateMachine() {

	for {
		select {
		case floorReached := <-eventReachedNewFloor:
			handleEventReachedNewFloor(floorReached)
		case <-eventCloseDoor:
			handleEventCloseDoor()
		case <-eventStop:
			// Not handled
		case <-eventObstruction:
			// Not handled
		case button := <-eventButtonFloorPressed:
			handleEventButtonPressed(button)
		}
	}
}

//-----------------------------------------------//

func Run() {

	err := elevator.Initialize()

	if err != nil{
		log.Error(err)
	}

	log.Warning("asd")
	currentState = STATE_STARTUP
	elevator.DriveInDirection(elevator.DIRECTION_DOWN)

	//go elevator.RegisterEvents(eventReachedNewFloor,
	//	eventStop,
	//	eventObstruction,
	//	eventButtonFloorPressed)

	//go stateMachine()
}
