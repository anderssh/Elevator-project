package elevator;

import(
	"../io"
	"time"
	"../log"
);

//-----------------------------------------------//

type Direction int;

const(
	DIRECTION_UP 	Direction = iota
	DIRECTION_DOWN 	Direction = iota
);

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

func RegisterEvents(eventReachedNewFloor chan int) {
	
	for {

		if io.ReadBit(io.SENSOR_FLOOR1) == 1 {
			eventReachedNewFloor <- 1;
		} else if io.ReadBit(io.SENSOR_FLOOR2) == 1 {
			eventReachedNewFloor <- 2;
		} else if io.ReadBit(io.SENSOR_FLOOR3) == 1 {
			eventReachedNewFloor <- 3;
		} else if io.ReadBit(io.SENSOR_FLOOR4) == 1 {
			eventReachedNewFloor <- 4;
		}

		time.Sleep(time.Microsecond*10);
	}
}

var eventReachedNewFloor chan int;

func Run() {

	go RegisterEvents(eventReachedNewFloor);
	
	DriveInDirection(DIRECTION_UP);

	for {
		select {
			case floor := <- eventReachedNewFloor:
				log.Data(floor);
				Stop();
		}
	}
}