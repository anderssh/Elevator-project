package elevator

import (
	"../io"
	"time"
	"../log"
)

//-----------------------------------------------//

type Direction int
type ButtonType int

const (
	DIRECTION_UP   		Direction = iota
	DIRECTION_DOWN 		Direction = iota

	BUTTON_CALL_UP     ButtonType = iota
	BUTTON_CALL_DOWN   ButtonType = iota
	BUTTON_CALL_INSIDE ButtonType = iota

	BUTTON_STOP        ButtonType = iota
	BUTTON_OBSTRUCTION ButtonType = iota

	NUMBER_OF_FLOORS int = 4
)

//-----------------------------------------------//

type ErrorElevator struct {
    errorMessage string
}

func (e *ErrorElevator) Error() string {
    return "Elevator error: " + e.errorMessage
}

//-----------------------------------------------//

type ButtonFloor struct {
	Type       ButtonType
	Floor      int
	Pressed    bool
	BusChannel int
}

type ButtonSimple struct {
	Type       ButtonType
	Pressed    bool
	BusChannel int
}

//-----------------------------------------------//

var numberOfFloors int = 4

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

func Initialize() *ErrorElevator {

	err := io.Initialize();
	log.Data("up")
	io.SetBit(io.LIGHT_FLOOR_IND2)
	if err != nil {
		log.Error(err)
		return &ErrorElevator{"Failed to initialize hardware."};
	}

	buttonStop 			= ButtonSimple{BUTTON_STOP, 		false, io.STOP}
	buttonObstruction 	= ButtonSimple{BUTTON_OBSTRUCTION, 	false, io.OBSTRUCTION}

	panel = make([]ButtonFloor, (NUMBER_OF_FLOORS*3)-2)

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		1, false, io.BUTTON_UP1})
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	1, false, io.BUTTON_COMMAND1})

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		2, false, io.BUTTON_UP2})
	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	2, false, io.BUTTON_DOWN2})
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	2, false, io.BUTTON_COMMAND2})

	panel = append(panel, ButtonFloor{BUTTON_CALL_UP, 		3, false, io.BUTTON_UP3})
	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	3, false, io.BUTTON_DOWN3})
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	3, false, io.BUTTON_COMMAND3})

	panel = append(panel, ButtonFloor{BUTTON_CALL_DOWN, 	4, false, io.BUTTON_DOWN4})
	panel = append(panel, ButtonFloor{BUTTON_CALL_INSIDE, 	4, false, io.BUTTON_COMMAND4})

	return nil;
}

//-----------------------------------------------//

func registerEventFloorReached(eventReachedNewFloor chan int) {

	if io.IsBitSet(io.SENSOR_FLOOR1) {
		eventReachedNewFloor <- 1
	} else if io.IsBitSet(io.SENSOR_FLOOR2) {
		eventReachedNewFloor <- 2
	} else if io.IsBitSet(io.SENSOR_FLOOR3) {
		eventReachedNewFloor <- 3
	} else if io.IsBitSet(io.SENSOR_FLOOR4) {
		eventReachedNewFloor <- 4
	}
}

//-----------------------------------------------//

func registerEventStop(eventStop chan bool) {

	buttonStopPreviouslyPressed := buttonStop.Pressed
	buttonStop.Pressed = io.IsBitSet(buttonStop.BusChannel)

	if buttonStop.Pressed && !buttonStopPreviouslyPressed {
		eventStop <- true
	}
}

//-----------------------------------------------//

func registerEventObstruction(eventObstruction chan bool) {

	buttonObstructionPreviouslyPressed := buttonObstruction.Pressed
	buttonObstruction.Pressed = io.IsBitSet(buttonObstruction.BusChannel)

	if buttonObstruction.Pressed && !buttonObstructionPreviouslyPressed {
		eventObstruction <- true
	} else if !buttonObstruction.Pressed && buttonObstructionPreviouslyPressed {
		eventObstruction <- false
	}
}

//-----------------------------------------------//

func registerEventButtonFloorPressed(eventButtonFloorPressed chan ButtonFloor) {

	for buttonIndex := range panel {

		buttonPreviouslyPressed := panel[buttonIndex].Pressed
		panel[buttonIndex].Pressed = io.IsBitSet(panel[buttonIndex].BusChannel)

		if panel[buttonIndex].Pressed && !buttonPreviouslyPressed {
			eventButtonFloorPressed <- panel[buttonIndex]
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
