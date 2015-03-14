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

var numberOfFloors 			int 	= 4;
var previouslyReachedFloor 	int 	= -1;

var buttonStop 			ButtonSimple
var buttonObstruction 	ButtonSimple

var containerButtonFloor []ButtonFloor

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

func initializeContainerButtonFloor() {
	
	containerButtonFloor = make([]ButtonFloor, (NUMBER_OF_FLOORS*3)-2);

	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_UP, 		1, false, false, io.BUTTON_UP1, io.LIGHT_UP1});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_INSIDE, 	1, false, false, io.BUTTON_CALL_INSIDE1, io.LIGHT_CALL_INSIDE1});

	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_UP, 		2, false, false, io.BUTTON_UP2, io.LIGHT_UP2});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_DOWN, 	2, false, false, io.BUTTON_DOWN2, io.LIGHT_DOWN2});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_INSIDE, 	2, false, false, io.BUTTON_CALL_INSIDE2, io.LIGHT_CALL_INSIDE2});

	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_UP, 		3, false, false, io.BUTTON_UP3, io.LIGHT_UP3});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_DOWN, 	3, false, false, io.BUTTON_DOWN3, io.LIGHT_DOWN3});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_INSIDE, 	3, false, false, io.BUTTON_CALL_INSIDE3, io.LIGHT_CALL_INSIDE3});

	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_DOWN, 	4, false, false, io.BUTTON_DOWN4, io.LIGHT_DOWN4});
	containerButtonFloor = append(containerButtonFloor, ButtonFloor{BUTTON_CALL_INSIDE, 	4, false, false, io.BUTTON_CALL_INSIDE4, io.LIGHT_CALL_INSIDE4});
}

func initializeSimpleBottons() {

	buttonStop 			= ButtonSimple{BUTTON_STOP, 		false, false, io.STOP};
	buttonObstruction 	= ButtonSimple{BUTTON_OBSTRUCTION, 	false, false, io.OBSTRUCTION};

}

func Initialize() *ErrorElevator {

	err := io.Initialize();

	if err != nil {
		log.Error(err);
		return &ErrorElevator{"Failed to initialize hardware."};
	}

	initializeContainerButtonFloor();
	initializeSimpleBottons();

	return nil;
}

//-----------------------------------------------//

func setFloorIndicatorLight(floorReached int){

	bit_1 :=  (floorReached - 1) 		% 2;
	bit_2 := ((floorReached - 1) >> 1)  % 2;
	
	log.Data("..")
	log.Data(floorReached)
	log.Data(bit_1)
	log.Data(bit_2)

	if bit_2 == 1 {
		io.SetBit(io.LIGHT_FLOOR_INDICATOR1)
	} else {
		io.ClearBit(io.LIGHT_FLOOR_INDICATOR1)
	}

	if bit_1 == 1 {
		io.SetBit(io.LIGHT_FLOOR_INDICATOR2)
	} else {
		io.ClearBit(io.LIGHT_FLOOR_INDICATOR2)
	}	
}

//-----------------------------------------------//

func SetPreviouslyReachedFloor(floorReached int){
	
	previouslyReachedFloor = floorReached;
	setFloorIndicatorLight(floorReached);

}

func GetPreviouslyReachedFloor() int{
	return previouslyReachedFloor;
}

func registerEventReachedFloor(eventReachedNewFloor chan int) {

	if io.IsBitSet(io.SENSOR_FLOOR1) {

		if previouslyReachedFloor != 1 {
			eventReachedNewFloor  <- 1;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR2) {
	
		if previouslyReachedFloor != 2 {
			eventReachedNewFloor  <- 2;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR3) {

		if previouslyReachedFloor != 3 {
			eventReachedNewFloor  <- 3;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR4) {
		
		if previouslyReachedFloor != 4 {
			eventReachedNewFloor  <- 4;
		}
	}
}

//-----------------------------------------------//

func registerEventStop(eventStop chan bool) {

	buttonStopPreviouslyPressed := buttonStop.Pressed;
	buttonStop.Pressed = io.IsBitSet(buttonStop.BusChannelPressed);

	if buttonStop.Pressed && !buttonStopPreviouslyPressed {
		eventStop <- true
	}
}

//-----------------------------------------------//

func registerEventObstruction(eventObstruction chan bool) {

	buttonObstructionPreviouslyPressed := buttonObstruction.Pressed;
	buttonObstruction.Pressed = io.IsBitSet(buttonObstruction.BusChannelPressed);

	if buttonObstruction.Pressed && !buttonObstructionPreviouslyPressed {
		eventObstruction <- true;
	} else if !buttonObstruction.Pressed && buttonObstructionPreviouslyPressed {
		eventObstruction <- false;
	}
}

//-----------------------------------------------//

func registerEventButtonFloorPressed(eventButtonFloorPressed chan ButtonFloor) {

	for buttonIndex := range containerButtonFloor {

		buttonPreviouslyPressed := containerButtonFloor[buttonIndex].Pressed;
		containerButtonFloor[buttonIndex].Pressed = io.IsBitSet(containerButtonFloor[buttonIndex].BusChannelPressed);

		if containerButtonFloor[buttonIndex].Pressed && !buttonPreviouslyPressed {
			eventButtonFloorPressed <- containerButtonFloor[buttonIndex];
		}
	}
}

//-----------------------------------------------//

func RegisterEvents(eventReachedNewFloor 	chan int,
					eventStop 				chan bool,
					eventObstruction 		chan bool,
					eventButtonFloorPressed chan ButtonFloor) {

	for {
		registerEventReachedFloor(eventReachedNewFloor)
		registerEventStop(eventStop)
		registerEventObstruction(eventObstruction)
		registerEventButtonFloorPressed(eventButtonFloorPressed)

		time.Sleep(time.Microsecond * 50)
	}
}
