package elevator

import (
	"../io"
	"time"
	"../log"
	. "../typeDefinitions"
)

//-----------------------------------------------//

type Direction int

const (
	DIRECTION_UP   		Direction 	= iota
	DIRECTION_DOWN 		Direction 	= iota

	NUMBER_OF_FLOORS 	int 		= 4
)

//-----------------------------------------------//

type ErrorElevator struct {
    errorMessage string
}

func (err *ErrorElevator) Error() string {
    return "Elevator error: " + err.errorMessage
}

//-----------------------------------------------//

var numberOfFloors 		int 			= 4;

var buttonStop 			ButtonSimple
var buttonObstruction 	ButtonSimple

var panel []ButtonFloor

//-----------------------------------------------//

func DriveInDirection(direction Direction) {

	if direction == DIRECTION_DOWN {
		io.SetBit(io.MOTORDIR)
		io.WriteAnalog(io.MOTOR, 2800)
	} else {
		io.ClearBit(io.MOTORDIR)
		io.WriteAnalog(io.MOTOR, 2800)
	}
}

func Stop() {
	io.WriteAnalog(io.MOTOR, 0)
}

//-----------------------------------------------//

func initializePanel() {
	
	panel = make([]ButtonFloor, (NUMBER_OF_FLOORS*3)-2);

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		1, false, io.BUTTON_UP1});
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	1, false, io.BUTTON_COMMAND1});

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		2, false, io.BUTTON_UP2});
	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	2, false, io.BUTTON_DOWN2});
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	2, false, io.BUTTON_COMMAND2});

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		3, false, io.BUTTON_UP3});
	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	3, false, io.BUTTON_DOWN3});
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	3, false, io.BUTTON_COMMAND3});

	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	4, false, io.BUTTON_DOWN4});
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	4, false, io.BUTTON_COMMAND4});
}

func Initialize() *ErrorElevator {

	err := io.Initialize();

	if err != nil {
		log.Error(err);
		return &ErrorElevator{"Failed to initialize hardware."};
	}

	buttonStop 			= ButtonSimple{BUTTON_STOP, 		false, io.STOP};
	buttonObstruction 	= ButtonSimple{BUTTON_OBSTRUCTION, 	false, io.OBSTRUCTION};

	initializePanel();

	return nil;
}

//-----------------------------------------------//

var previouslyReachedFloor int = -1;

func registerEventFloorReached(eventReachedNewFloor chan int) {

	if io.IsBitSet(io.SENSOR_FLOOR1) {

		if previouslyReachedFloor != 1 {
			eventReachedNewFloor  <- 1;
			previouslyReachedFloor = 1;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR2) {
	
		if previouslyReachedFloor != 2 {
			eventReachedNewFloor  <- 2;
			previouslyReachedFloor = 2;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR3) {

		if previouslyReachedFloor != 3 {
			eventReachedNewFloor  <- 3;
			previouslyReachedFloor = 3;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR4) {
		
		if previouslyReachedFloor != 4 {
			eventReachedNewFloor  <- 4;
			previouslyReachedFloor = 4;
		}
	}
}

//-----------------------------------------------//

func registerEventStop(eventStop chan bool) {

	buttonStopPreviouslyPressed := buttonStop.Pressed;
	buttonStop.Pressed = io.IsBitSet(buttonStop.BusChannel);

	if buttonStop.Pressed && !buttonStopPreviouslyPressed {
		eventStop <- true
	}
}

//-----------------------------------------------//

func registerEventObstruction(eventObstruction chan bool) {

	buttonObstructionPreviouslyPressed := buttonObstruction.Pressed;
	buttonObstruction.Pressed = io.IsBitSet(buttonObstruction.BusChannel);

	if buttonObstruction.Pressed && !buttonObstructionPreviouslyPressed {
		eventObstruction <- true;
	} else if !buttonObstruction.Pressed && buttonObstructionPreviouslyPressed {
		eventObstruction <- false;
	}
}

//-----------------------------------------------//

func registerEventButtonFloorPressed(eventButtonFloorPressed chan ButtonFloor) {

	for buttonIndex := range panel {

		buttonPreviouslyPressed := panel[buttonIndex].Pressed;
		panel[buttonIndex].Pressed = io.IsBitSet(panel[buttonIndex].BusChannel);

		if panel[buttonIndex].Pressed && !buttonPreviouslyPressed {
			eventButtonFloorPressed <- panel[buttonIndex];
		}
	}
}

//-----------------------------------------------//

func RegisterEvents(eventReachedNewFloor 	chan int,
					eventStop 				chan bool,
					eventObstruction 		chan bool,
					eventButtonFloorPressed chan ButtonFloor) {

	for {
		registerEventFloorReached(eventReachedNewFloor)
		registerEventStop(eventStop)
		registerEventObstruction(eventObstruction)
		registerEventButtonFloorPressed(eventButtonFloorPressed)

		time.Sleep(time.Microsecond * 50)
	}
}
