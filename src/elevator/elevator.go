package elevator;

import(
	"../io"
	"time"
);

//-----------------------------------------------//

type Direction 	int;
type ButtonType int;

const(
	DIRECTION_UP 		Direction 	= iota
	DIRECTION_DOWN 		Direction 	= iota

	BUTTON_CALL_UP  	ButtonType  = iota
	BUTTON_CALL_DOWN  	ButtonType  = iota
	BUTTON_CALL_INSIDE  ButtonType  = iota
);

type Button struct {
	Type 		ButtonType
	Floor 		int
}

//-----------------------------------------------//

func DriveInDirection(direction Direction) {
	if direction == DIRECTION_DOWN {
		io.SetBit(io.MOTORDIR);
		io.WriteAnalog(io.MOTOR, 2800);
	} else {
		io.ClearBit(io.MOTORDIR);
		io.WriteAnalog(io.MOTOR, 2800);
	}
}

func Stop() {
	io.WriteAnalog(io.MOTOR, 0);
}

//-----------------------------------------------//

func Initialize() bool {

	if !io.Initialize() {
		return false;
	}

	return true;
}

//-----------------------------------------------//

func registerEventFloorReached(eventReachedNewFloor chan int) {
	
	if io.IsBitSet(io.SENSOR_FLOOR1) {
		eventReachedNewFloor <- 1;
	} else if io.IsBitSet(io.SENSOR_FLOOR2) {
		eventReachedNewFloor <- 2;
	} else if io.IsBitSet(io.SENSOR_FLOOR3) {
		eventReachedNewFloor <- 3;
	} else if io.IsBitSet(io.SENSOR_FLOOR4) {
		eventReachedNewFloor <- 4;
	}
}

func registerEventStop(eventStop chan bool) {
		
}

func registerEventObstruction(eventObstruction chan bool) {
		
}

func registerEventButtonPressed(eventButtonPressed chan Button) {
		
}



//-----------------------------------------------//

func RegisterEvents(eventReachedNewFloor chan int, eventStop chan bool, eventObstruction chan bool, eventButtonPressed chan Button) {
		
	for {
		registerEventFloorReached(eventReachedNewFloor);
		registerEventStop(eventStop);
		registerEventObstruction(eventObstruction);
		registerEventButtonPressed(eventButtonPressed);
		
		time.Sleep(time.Microsecond*50);
	}
}