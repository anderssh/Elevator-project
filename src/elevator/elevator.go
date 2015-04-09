package elevator

import (
	"../io"
	"time"
	"../log"
	. "../typeDefinitions"
)

//-----------------------------------------------//

const (
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
var lastReachedFloor 		int 	= -1;

var buttonStop 			ButtonSimple
var buttonObstruction 	ButtonSimple

var containerButtonFloor [][]ButtonFloor

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
		containerButtonFloor[i] = make([]ButtonFloor, NUMBER_OF_FLOORS)
	}
	containerButtonFloor[0][0] = ButtonFloor{BUTTON_CALL_UP, 		0, false, false, io.BUTTON_UP0, io.LIGHT_UP0};
	containerButtonFloor[0][1] = ButtonFloor{BUTTON_CALL_UP, 		1, false, false, io.BUTTON_UP1, io.LIGHT_UP1};
	containerButtonFloor[0][2] = ButtonFloor{BUTTON_CALL_UP, 		2, false, false, io.BUTTON_UP2, io.LIGHT_UP2};
	containerButtonFloor[0][3] = ButtonFloor{};
	
	containerButtonFloor[1][0] = ButtonFloor{};
	containerButtonFloor[1][1] = ButtonFloor{BUTTON_CALL_DOWN, 	1, false, false, io.BUTTON_DOWN1, io.LIGHT_DOWN1};
	containerButtonFloor[1][2] = ButtonFloor{BUTTON_CALL_DOWN, 	2, false, false, io.BUTTON_DOWN2, io.LIGHT_DOWN2};
	containerButtonFloor[1][3] = ButtonFloor{BUTTON_CALL_DOWN, 	3, false, false, io.BUTTON_DOWN3, io.LIGHT_DOWN3};
	
	containerButtonFloor[2][0] = ButtonFloor{BUTTON_CALL_INSIDE, 	0, false, false, io.BUTTON_CALL_INSIDE0, io.LIGHT_CALL_INSIDE0};
	containerButtonFloor[2][1] = ButtonFloor{BUTTON_CALL_INSIDE, 	1, false, false, io.BUTTON_CALL_INSIDE1, io.LIGHT_CALL_INSIDE1};
	containerButtonFloor[2][2] = ButtonFloor{BUTTON_CALL_INSIDE, 	2, false, false, io.BUTTON_CALL_INSIDE2, io.LIGHT_CALL_INSIDE2};
	containerButtonFloor[2][3] = ButtonFloor{BUTTON_CALL_INSIDE, 	3, false, false, io.BUTTON_CALL_INSIDE3, io.LIGHT_CALL_INSIDE3};
}
func initializeSimpleButtons() {

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
	initializeSimpleButtons();
	for floor := 0; floor < NUMBER_OF_FLOORS ; floor++ {
	
		TurnOffAllLightButtonsOnFloor(floor)
	}
	return nil;
}

//-----------------------------------------------//

func switchLightFloorIndicator(floorReached int) {

	bit_1 :=  (floorReached) 		% 2;
	bit_2 := ((floorReached) >> 1)  % 2;

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

func TurnOnLightDoorOpen() {
	io.SetBit(io.LIGHT_DOOR_OPEN)
}

func TurnOffLightDoorOpen() {
	io.ClearBit(io.LIGHT_DOOR_OPEN)
}

func TurnOnLightButtonFromOrder(order Order) {

	if order.Type == ORDER_UP{
		io.SetBit(containerButtonFloor[0][order.Floor].BusChannelLight)
	} else if order.Type == ORDER_DOWN{
		io.SetBit(containerButtonFloor[1][order.Floor].BusChannelLight)
	} else if order.Type == ORDER_INSIDE{
		io.SetBit(containerButtonFloor[2][order.Floor].BusChannelLight)
	}
}

func TurnOffLightButtonFromOrder(order Order) {

	if order.Type == ORDER_UP{
		io.ClearBit(containerButtonFloor[0][order.Floor].BusChannelLight)
	} else if order.Type == ORDER_DOWN{
		io.ClearBit(containerButtonFloor[1][order.Floor].BusChannelLight)
	} else if order.Type == ORDER_INSIDE{
		io.ClearBit(containerButtonFloor[2][order.Floor].BusChannelLight)
	}
}

func TurnOffAllLightButtonsOnFloor(floor int) {
	io.ClearBit(containerButtonFloor[0][floor].BusChannelLight)
	io.ClearBit(containerButtonFloor[1][floor].BusChannelLight)
	io.ClearBit(containerButtonFloor[2][floor].BusChannelLight)
}
//-----------------------------------------------//

func SetLastReachedFloor(floorReached int) {
	
	lastReachedFloor = floorReached;
	switchLightFloorIndicator(floorReached);

}

func GetLastReachedFloor() int {
	return lastReachedFloor;
}

func registerEventReachedFloor(eventReachedNewFloor chan int) {

	if io.IsBitSet(io.SENSOR_FLOOR0) {

		if lastReachedFloor != 0{
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

	for buttonTypeIndex := range containerButtonFloor {
		for buttonFloorIndex := range containerButtonFloor[buttonTypeIndex] {

			button := containerButtonFloor[buttonTypeIndex][buttonFloorIndex];

			if !(button.Type == BUTTON_CALL_DOWN && button.Floor == 0) && !(button.Type == BUTTON_CALL_UP && button.Floor == NUMBER_OF_FLOORS - 1) { 		// Omitting the non-existing buttons

				buttonPreviouslyPressed := button.Pressed;
				button.Pressed 			 = io.IsBitSet(button.BusChannelPressed);

				containerButtonFloor[buttonTypeIndex][buttonFloorIndex] = button; // Workaround for go bug

				if button.Pressed && !buttonPreviouslyPressed {
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
