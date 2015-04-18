package elevator

import (
	"user/io"
	"time"
	. "user/typeDefinitions"
	"user/log"
);

//-----------------------------------------------//

const(
	NUMBER_OF_FLOORS int = 4
);

//-----------------------------------------------//

type ErrorElevator struct {
    errorMessage string
}

func (err *ErrorElevator) Error() string {
    return "Elevator error: " + err.errorMessage
}

//-----------------------------------------------//

var lastReachedFloor 	int;

var buttonStop 			ButtonSimple;
var buttonObstruction 	ButtonSimple;

var containerButtonFloor [][]ButtonFloor;

//-----------------------------------------------//

var direction Direction;

func GetDirection() Direction {
	return direction;
}

//-----------------------------------------------//

func DriveInDirection(requestedDirection Direction) {

	if requestedDirection == DIRECTION_DOWN {

		io.SetBit(io.MOTORDIR);
		io.WriteAnalog(io.MOTOR, 2800);

	} else {

		io.ClearBit(io.MOTORDIR);
		io.WriteAnalog(io.MOTOR, 2800);
	}

	direction = requestedDirection;
}

func Stop() {
	io.WriteAnalog(io.MOTOR, 0);
}

//-----------------------------------------------//

func initializeContainerButtonFloor() {
	
	containerButtonFloor = make([][]ButtonFloor, 3);

	for i := range containerButtonFloor {
		containerButtonFloor[i] = make([]ButtonFloor, NUMBER_OF_FLOORS);
	}

	containerButtonFloor[0][0] = ButtonFloor{ Type : BUTTON_CALL_UP, Floor : 0, IoRegisterPressed : io.BUTTON_UP0, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_UP0 };
	containerButtonFloor[0][1] = ButtonFloor{ Type : BUTTON_CALL_UP, Floor : 1, IoRegisterPressed : io.BUTTON_UP1, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_UP1 };
	containerButtonFloor[0][2] = ButtonFloor{ Type : BUTTON_CALL_UP, Floor : 2, IoRegisterPressed : io.BUTTON_UP2, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_UP2 };
	containerButtonFloor[0][3] = ButtonFloor{ Type : BUTTON_CALL_UP, Floor : 3, IoRegisterPressed : io.BUTTON_UP3, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_UP3 };

	containerButtonFloor[1][0] = ButtonFloor{ Type : BUTTON_CALL_DOWN, Floor : 0, IoRegisterPressed : io.BUTTON_DOWN0, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_DOWN0 };
	containerButtonFloor[1][1] = ButtonFloor{ Type : BUTTON_CALL_DOWN, Floor : 1, IoRegisterPressed : io.BUTTON_DOWN1, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_DOWN1 };
	containerButtonFloor[1][2] = ButtonFloor{ Type : BUTTON_CALL_DOWN, Floor : 2, IoRegisterPressed : io.BUTTON_DOWN2, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_DOWN2 };
	containerButtonFloor[1][3] = ButtonFloor{ Type : BUTTON_CALL_DOWN, Floor : 3, IoRegisterPressed : io.BUTTON_DOWN3, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_DOWN3 };
	
	containerButtonFloor[2][0] = ButtonFloor{ Type : BUTTON_CALL_INSIDE, Floor : 0, IoRegisterPressed : io.BUTTON_CALL_INSIDE0, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_CALL_INSIDE0 };
	containerButtonFloor[2][1] = ButtonFloor{ Type : BUTTON_CALL_INSIDE, Floor : 1, IoRegisterPressed : io.BUTTON_CALL_INSIDE1, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_CALL_INSIDE1 };
	containerButtonFloor[2][2] = ButtonFloor{ Type : BUTTON_CALL_INSIDE, Floor : 2, IoRegisterPressed : io.BUTTON_CALL_INSIDE2, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_CALL_INSIDE2 };
	containerButtonFloor[2][3] = ButtonFloor{ Type : BUTTON_CALL_INSIDE, Floor : 3, IoRegisterPressed : io.BUTTON_CALL_INSIDE3, PressedReadingPrevious : false, PressedReadingCurrent : false, IoRegisterLight : io.LIGHT_CALL_INSIDE3 };
}

func initializeSimpleButtons() {

	buttonStop 			= ButtonSimple{ Type : BUTTON_STOP, 		IoRegisterPressed : io.STOP, 		PressedReadingPrevious : false, PressedReadingCurrent : false };
	buttonObstruction 	= ButtonSimple{ Type : BUTTON_OBSTRUCTION, 	IoRegisterPressed : io.OBSTRUCTION, PressedReadingPrevious : false, PressedReadingCurrent : false };
}

func Initialize() *ErrorElevator {

	err := io.Initialize();

	if err != nil {
		log.Error(err);
		return &ErrorElevator{"Failed to initialize hardware."};
	}

	lastReachedFloor = -1;

	initializeContainerButtonFloor();
	initializeSimpleButtons();

	for floor := 0; floor < NUMBER_OF_FLOORS; floor++ {
		TurnOffAllLightButtonsOnFloor(floor)
	}

	return nil;
}

//-----------------------------------------------//

func TurnOnLightDoorOpen() {
	io.SetBit(io.LIGHT_DOOR_OPEN);
}

func TurnOffLightDoorOpen() {
	io.ClearBit(io.LIGHT_DOOR_OPEN);
}

//-----------------------------------------------//

func TurnOnLightButtonFromOrder(order Order) {

	if order.Type == ORDER_UP{
		containerButtonFloor[0][order.Floor].TurnOnLight();
	} else if order.Type == ORDER_DOWN{
		containerButtonFloor[1][order.Floor].TurnOnLight();
	} else if order.Type == ORDER_INSIDE{
		containerButtonFloor[2][order.Floor].TurnOnLight();
	}
}

func TurnOffLightButtonFromOrder(order Order) {

	if order.Type == ORDER_UP{
		containerButtonFloor[0][order.Floor].TurnOffLight();
	} else if order.Type == ORDER_DOWN{
		containerButtonFloor[1][order.Floor].TurnOffLight();
	} else if order.Type == ORDER_INSIDE{
		containerButtonFloor[2][order.Floor].TurnOffLight();
	}
}

func TurnOffAllLightButtonsOnFloor(floor int) {

	if floor < NUMBER_OF_FLOORS - 1 {
		containerButtonFloor[0][floor].TurnOffLight(); 		// ORDER_UP
	}

	if floor > 0 {
		containerButtonFloor[1][floor].TurnOffLight(); 		// ORDER_DOWN
	}

	containerButtonFloor[2][floor].TurnOffLight();
}

//-----------------------------------------------//

func switchLightFloorIndicator(floorReached int) {

	// 00: Floor 0
	// 01: Floor 1
	// 10: Floor 2
	// 11: Floor 3

	bit_1 :=  (floorReached) 		% 2;
	bit_2 := ((floorReached) >> 1)  % 2;

	if bit_2 == 1 {
		io.SetBit(io.LIGHT_FLOOR_INDICATOR1);
	} else {
		io.ClearBit(io.LIGHT_FLOOR_INDICATOR1);
	}

	if bit_1 == 1 {
		io.SetBit(io.LIGHT_FLOOR_INDICATOR2);
	} else {
		io.ClearBit(io.LIGHT_FLOOR_INDICATOR2);
	}
}

func SetLastReachedFloor(floorReached int) {
	
	lastReachedFloor = floorReached;
	switchLightFloorIndicator(floorReached);
}

func GetLastReachedFloor() int {
	return lastReachedFloor;
}

func registerEventReachedFloor(eventReachedNewFloor chan int) {

	if io.IsBitSet(io.SENSOR_FLOOR0) {

		if lastReachedFloor != 0 {
			eventReachedNewFloor  <- 0;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR1) {
	
		if lastReachedFloor != 1 {
			eventReachedNewFloor  <- 1;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR2) {

		if lastReachedFloor != 2 {
			eventReachedNewFloor  <- 2;
		}

	} else if io.IsBitSet(io.SENSOR_FLOOR3) {
		
		if lastReachedFloor != 3 {
			eventReachedNewFloor  <- 3;
		}
	}
}

//-----------------------------------------------//

func registerEventStop(eventStop chan bool) {

	buttonStop.UpdateState();

	if buttonStop.IsPressed() {
		eventStop <- true;
	}
}

//-----------------------------------------------//

func registerEventObstruction(eventObstruction chan bool) {

	buttonObstruction.UpdateState();

	if buttonObstruction.IsPressed() {
		eventObstruction <- true;
	} else if buttonObstruction.IsReleased() {
		eventObstruction <- false;
	}
}

//-----------------------------------------------//

func registerEventButtonFloorPressed(eventButtonFloorPressed chan ButtonFloor) {

	for buttonTypeIndex := range containerButtonFloor {
		for buttonFloorIndex := range containerButtonFloor[buttonTypeIndex] {

			button := containerButtonFloor[buttonTypeIndex][buttonFloorIndex];

			if !(button.Type == BUTTON_CALL_DOWN && button.Floor == 0) && !(button.Type == BUTTON_CALL_UP && button.Floor == NUMBER_OF_FLOORS - 1) { // Omit non existing buttons

				button.UpdateState(); containerButtonFloor[buttonTypeIndex][buttonFloorIndex] = button; // Workaround for GO bug

				if button.IsPressed() {
					eventButtonFloorPressed <- button;
				}
			}
		}
	}
}

//-----------------------------------------------//

func RegisterEvents(eventReachedNewFloor 	chan int,
					eventStop 				chan bool,
					eventObstruction 		chan bool,
					eventButtonFloorPressed chan ButtonFloor) {

	for {
		registerEventReachedFloor(eventReachedNewFloor);
		registerEventStop(eventStop);
		registerEventObstruction(eventObstruction);
		registerEventButtonFloorPressed(eventButtonFloorPressed);

		time.Sleep(time.Microsecond * 50);
	}
}
