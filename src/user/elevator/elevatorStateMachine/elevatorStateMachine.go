package elevatorStateMachine

import(
	. "user/typeDefinitions"
	"user/config"
	"user/elevator/elevatorObject"
	"user/log"
	"user/orders/ordersLocal"
	"time"
	"fmt"
	"user/network"
	"user/encoder/JSON"
	"os"
	"os/signal"
	"io/ioutil"
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

//-----------------------------------------------//

func backupOrdersLocalToFile(backupOrdersLocal chan []OrderLocal) {

	for {
		ordersToWrite := <- backupOrdersLocal;

		ordersToWriteEncoded, _ := JSON.Encode(ordersToWrite);

		ioutil.WriteFile(config.BACKUP_FILE_NAME, ordersToWriteEncoded, 0644); // 0644: for read/write permissions
	}
}

//-----------------------------------------------//

func Display() {
	
	if config.SHOULD_DISPLAY_ELEVATOR {

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

		for floor := 3; floor >= 0; floor-- {

			//-----------------------------------------------//
			// Elevator position

			fmt.Print(floor);
			fmt.Print("|"); // Left wall

			if elevatorObject.GetLastReachedFloor() == floor {
				fmt.Print("\x1b[33;1m");
				fmt.Print("+++");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("   ");
			}

			fmt.Print("|"); // Right wall

			if currentState == STATE_DOOR_OPEN && elevatorObject.GetLastReachedFloor() == floor {
				fmt.Print("->");
			} else {
				fmt.Print("  ");
			}

			//-----------------------------------------------//
			// Order display

			fmt.Print("\t");

			if ordersLocal.AlreadyStored(OrderLocal{ Type : ORDER_INSIDE, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("O");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("O");
			}

			fmt.Print(" ");

			if ordersLocal.AlreadyStored(OrderLocal{ Type : ORDER_UP, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("^");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("^");
			}

			fmt.Print(" ");

			if ordersLocal.AlreadyStored(OrderLocal{ Type : ORDER_DOWN, Floor : floor }) {
				fmt.Print("\x1b[31;1m");
				fmt.Print("_");
				fmt.Print("\x1b[0m");
			} else {
				fmt.Print("_");
			}

			fmt.Print("\n");
		}
	}
}

//-----------------------------------------------//

func handleReachedNewFloor(floorReached int, workerExitsStartup chan bool, eventCloseDoor chan bool, backupDataOrdersLocal []OrderLocal) {
	
	switch currentState {
		case STATE_STARTUP:

			elevatorObject.SetLastReachedFloor(floorReached);

			// Restore state from backup
			for orderIndex := range backupDataOrdersLocal {

				order := backupDataOrdersLocal[orderIndex];

				ordersLocal.Add(order, elevatorObject.GetLastReachedFloor(), false, elevatorObject.GetDirection());
				elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);
			}

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();

				if floorDestination == floorReached {
				
					elevatorObject.Stop();

					elevatorObject.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true
					});

					currentState = STATE_DOOR_OPEN;

				} else {

					if floorDestination > elevatorObject.GetLastReachedFloor() {
						elevatorObject.DriveInDirection(DIRECTION_UP);
					} else {
						elevatorObject.DriveInDirection(DIRECTION_DOWN);
					}

					currentState = STATE_MOVING;
				}

			} else {

				elevatorObject.Stop();
				currentState = STATE_IDLE;
			}
			// End restore from backup

			workerExitsStartup <- true;

		case STATE_IDLE:

			elevatorObject.SetLastReachedFloor(floorReached);

		case STATE_MOVING:
			
			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();

				if floorDestination == floorReached {
				
					elevatorObject.Stop();

					elevatorObject.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true
					});

					currentState = STATE_DOOR_OPEN;

				} else { // Update if other elevators has taken the orders

					if floorDestination > elevatorObject.GetLastReachedFloor() {
						elevatorObject.DriveInDirection(DIRECTION_UP);
					} else {
						elevatorObject.DriveInDirection(DIRECTION_DOWN);
					}
				}

			} else {

				elevatorObject.Stop();
				currentState = STATE_IDLE;
			}

			elevatorObject.SetLastReachedFloor(floorReached);

		case STATE_DOOR_OPEN:

			elevatorObject.SetLastReachedFloor(floorReached);
	}
}

//-----------------------------------------------//

func handleCloseDoor(eventCloseDoor chan bool, workerOrdersExecutedOnFloor chan int) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Closed door in state startup");

		case STATE_IDLE:

			log.Warning("Closed door in state idle");

		case STATE_MOVING:

			log.Warning("Closed door in state moving");

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveOnFloor(elevatorObject.GetLastReachedFloor());
			workerOrdersExecutedOnFloor <- elevatorObject.GetLastReachedFloor();
			
			elevatorObject.TurnOffLightDoorOpen();
			elevatorObject.TurnOffAllLightsOnButtonsOnFloor(elevatorObject.GetLastReachedFloor());

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();
			
				if floorDestination == elevatorObject.GetLastReachedFloor() {
					
					log.Warning("Orders still on floor when door closed.");
					
					elevatorObject.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true
					});

					currentState = STATE_DOOR_OPEN;

				} else {

					if floorDestination > elevatorObject.GetLastReachedFloor() {
						elevatorObject.DriveInDirection(DIRECTION_UP);
					} else {
						elevatorObject.DriveInDirection(DIRECTION_DOWN);
					}

					currentState = STATE_MOVING;
				}
			} else {
				currentState = STATE_IDLE;
			}
	}
}

//-----------------------------------------------//

func handleButtonPressed(button ButtonFloor, workerNewOrder chan OrderLocal) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to order at startup. Order not registered.");

		case STATE_IDLE:

			workerNewOrder <- button.ConvertToOrderLocal();

		case STATE_MOVING:

			workerNewOrder <- button.ConvertToOrderLocal();

		case STATE_DOOR_OPEN:

			workerNewOrder <- button.ConvertToOrderLocal();
	}
}

//-----------------------------------------------//

func handleNewDestinationOrder(order OrderLocal, eventCloseDoor chan bool) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle order at startup. Order not registered.");

		case STATE_IDLE:

			if !ordersLocal.AlreadyStored(order) {

				ordersLocal.Add(order, elevatorObject.GetLastReachedFloor(), false, elevatorObject.GetDirection());
				elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);
			}

			if ordersLocal.Exists() {

				floorDestination = ordersLocal.GetDestination();
				
				if (floorDestination == elevatorObject.GetLastReachedFloor()) {
					
					elevatorObject.TurnOnLightDoorOpen();
					time.AfterFunc(config.ELEVATOR_DOOR_OPEN_DURATION, func() { // Close the door
						eventCloseDoor <- true;
					});

					currentState = STATE_DOOR_OPEN;

				} else if floorDestination < elevatorObject.GetLastReachedFloor() {
					
					elevatorObject.DriveInDirection(DIRECTION_DOWN);
					currentState = STATE_MOVING;

				} else {

					elevatorObject.DriveInDirection(DIRECTION_UP);
					currentState = STATE_MOVING;
				}
			}

		case STATE_MOVING:

			if !ordersLocal.AlreadyStored(order) {
				
				ordersLocal.Add(order, elevatorObject.GetLastReachedFloor(), true, elevatorObject.GetDirection());
				elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);

				floorDestination = ordersLocal.GetDestination();
			}

		case STATE_DOOR_OPEN:

			if !ordersLocal.AlreadyStored(order) {
				
				ordersLocal.Add(order, elevatorObject.GetLastReachedFloor(), false, elevatorObject.GetDirection());
				elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);

				floorDestination = ordersLocal.GetDestination();
			}
	}
}

//-----------------------------------------------//

func handleDestinationOrderTakenBySomeone(order OrderLocal) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle order taken by someone at startup.");

		case STATE_IDLE:

			elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);

		case STATE_MOVING:

			elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);

		case STATE_DOOR_OPEN:

			elevatorObject.TurnOnLightOnButtonFromOrderLocal(order);
	}
}

//-----------------------------------------------//

func handleOrdersExectuedOnFloorBySomeone(floor int) {

	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to handle event orders executed by someone at startup.");

		case STATE_IDLE:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevatorObject.TurnOffCallUpAndCallDownLightsOnButtonsOnFloor(floor);

		case STATE_MOVING:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevatorObject.TurnOffCallUpAndCallDownLightsOnButtonsOnFloor(floor);

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveCallUpAndCallDownOnFloor(floor);
			elevatorObject.TurnOffCallUpAndCallDownLightsOnButtonsOnFloor(floor);
	}
}

//-----------------------------------------------//

func handleCostRequest(order OrderLocal, workerCostResponse chan int) {
	
	switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried to get cost when in startup. Nothing happens");

		case STATE_IDLE:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevatorObject.GetLastReachedFloor(), false, elevatorObject.GetDirection());

		case STATE_MOVING:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevatorObject.GetLastReachedFloor(), true, elevatorObject.GetDirection());

		case STATE_DOOR_OPEN:

			workerCostResponse <- ordersLocal.GetCostOf(order, elevatorObject.GetLastReachedFloor(), false, elevatorObject.GetDirection());
	}
}

//-----------------------------------------------//

func handleRemoveCallUpAndCallDownOrders() {

		switch currentState {
		case STATE_STARTUP:

			log.Warning("Tried remove call up and down orders at startup.");

		case STATE_IDLE:

			ordersLocal.RemoveCallUpAndCallDown();

		case STATE_MOVING:

			ordersLocal.RemoveCallUpAndCallDown();

		case STATE_DOOR_OPEN:

			ordersLocal.RemoveCallUpAndCallDown();
	}
}

//-----------------------------------------------//

func Run(transmitChannelUDP 					chan network.Message,

		 backupDataOrdersLocal 						[]OrderLocal,

		 eventNewDestinationOrder 				chan OrderLocal,
		 eventCostRequest 						chan OrderLocal,
		 eventOrdersExecutedOnFloorBySomeone 	chan int,
		 eventDestinationOrderTakenBySomeone 	chan OrderLocal,
		 eventRemoveCallUpAndCallDownOrders 	chan bool,

		 workerExitsStartup 					chan bool,
		 workerNewOrder 						chan OrderLocal,
		 workerCostResponse 					chan int,
		 workerOrdersExecutedOnFloor 			chan int) {

	//-----------------------------------------------//

	eventProgramTermination := make(chan os.Signal, 1);

	signal.Notify(eventProgramTermination, os.Interrupt);

	//-----------------------------------------------//

	backupOrdersLocal := make(chan []OrderLocal, 1000);

	go backupOrdersLocalToFile(backupOrdersLocal);

	//-----------------------------------------------//

	elevatorObject.Initialize();

	currentState 	 = STATE_STARTUP;
	floorDestination = -1;

	//-----------------------------------------------//

	eventReachedNewFloor 	:= make(chan int);
	eventCloseDoor 			:= make(chan bool);
	eventStop 				:= make(chan bool);
	eventObstruction 		:= make(chan bool);
	eventButtonFloorPressed := make(chan ButtonFloor);

	go elevatorObject.RegisterEvents(eventReachedNewFloor,
									 eventStop,
									 eventObstruction,
									 eventButtonFloorPressed);

	elevatorObject.DriveInDirection(DIRECTION_DOWN);

	//-----------------------------------------------//

	for {
		select {
			case <- eventStop:
				
				log.Warning("Pushed stop button. Button has no effect.");

			case <- eventObstruction:

				log.Warning("Pushed obstruction button. Button has no effect.");

			case floorReached := <- eventReachedNewFloor:

				handleReachedNewFloor(floorReached, workerExitsStartup, eventCloseDoor, backupDataOrdersLocal);
				backupOrdersLocal <- ordersLocal.MakeBackup();
				
				Display();

			case <- eventCloseDoor:

				handleCloseDoor(eventCloseDoor, workerOrdersExecutedOnFloor);
				backupOrdersLocal <- ordersLocal.MakeBackup();
				
				Display();

			case button := <- eventButtonFloorPressed:

				handleButtonPressed(button, workerNewOrder);
				
				Display();

			case order := <- eventNewDestinationOrder:

				handleNewDestinationOrder(order, eventCloseDoor);
				backupOrdersLocal <- ordersLocal.MakeBackup();
				
				Display();

			case order := <- eventDestinationOrderTakenBySomeone:

				handleDestinationOrderTakenBySomeone(order);
				
				Display();

			case floor := <- eventOrdersExecutedOnFloorBySomeone:

				handleOrdersExectuedOnFloorBySomeone(floor);
				backupOrdersLocal <- ordersLocal.MakeBackup();
				
				Display();

			case <- eventRemoveCallUpAndCallDownOrders:

				handleRemoveCallUpAndCallDownOrders();
				backupOrdersLocal <- ordersLocal.MakeBackup();

			case order := <- eventCostRequest:

				handleCostRequest(order, workerCostResponse);
				Display();

			case signal := <- eventProgramTermination:

				log.Error("Program terminated.", signal);
				elevatorObject.Stop();
				os.Exit(1);
		}
	}
}