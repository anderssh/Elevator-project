package ordersLocal;

import(
	. "user/typeDefinitions"
	"user/config"
	"math"
	"time"
);

//-----------------------------------------------//

const(
	COST_FOR_STOP float64 = 2
);

//-----------------------------------------------//

var orders []Order = make([]Order, 0, 1);

//-----------------------------------------------//

func MakeBackup() OrdersBackup {

	ordersToBackup := make([]Order, 0, 1); 				// Only backup inside order, other are distributed via ordersGlobal

	for orderIndex := range orders {
		if orders[orderIndex].Type == ORDER_INSIDE {
			ordersToBackup = append(ordersToBackup, orders[orderIndex]);
		}
	}

	return OrdersBackup{ Orders : ordersToBackup, Timestamp : time.Now().UnixNano() };
}

//-----------------------------------------------//

func Exists() bool {
	
	if len(orders) > 0{
		return true;
	}

	return false;
}

func AlreadyStored(order Order) bool {
	
	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type  && orders[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

//-----------------------------------------------//

func GetDestination() int {
	return orders[0].Floor;
}

//-----------------------------------------------//

func RemoveOnFloor(floor int) {
	
	orderIndex := 0;

	for {
		if orderIndex >= 0 && orderIndex < len(orders) {

			if (orders[orderIndex].Floor == floor) {

				orders = append(orders[0:orderIndex], orders[orderIndex + 1:] ... );

			} else {
				orderIndex = orderIndex + 1;
			}

		} else {
			break;
		}
	}
}

func RemoveCallUpAndCallDown() {
	
	orderIndex := 0;

	for {
		if orderIndex >= 0 && orderIndex < len(orders) {

			if !(orders[orderIndex].Type == ORDER_INSIDE) {

				orders = append(orders[0:orderIndex], orders[orderIndex + 1:] ... );

			} else {
				orderIndex = orderIndex + 1;
			}

		} else {
			break;
		}
	}
}

func RemoveCallUpAndCallDownOnFloor(floor int) {
	
	orderIndex := 0;

	for {
		if orderIndex >= 0 && orderIndex < len(orders) {

			if (orders[orderIndex].Floor == floor && !(orders[orderIndex].Type == ORDER_INSIDE)) {

				orders = append(orders[0:orderIndex], orders[orderIndex + 1:] ... );

			} else {
				orderIndex = orderIndex + 1;
			}

		} else {
			break;
		}
	}
}

//-----------------------------------------------//

func shouldOrderBeBetween(order Order, floorStart int, floorEnd int) bool {
		
	// In between if moving up
	if order.Type == ORDER_UP {

		floorLower := floorStart;
		floorUpper := floorEnd;

		if floorLower <= floorUpper && order.Floor >= floorLower && order.Floor <= floorUpper {
			return true;
		}

	// In between if moving down
	} else if order.Type == ORDER_DOWN {

		floorLower := floorEnd;
		floorUpper := floorStart;

		if floorLower <= floorUpper && order.Floor >= floorLower && order.Floor <= floorUpper {
			return true;
		}

	// In between is all that is needed, not dependent on moving direction
	} else if order.Type == ORDER_INSIDE {

		if floorStart <= floorEnd {

			floorLower := floorStart;
			floorUpper := floorEnd;

			if order.Floor >= floorLower && order.Floor <= floorUpper {
				return true;
			}

		} else {

			floorLower := floorEnd;
			floorUpper := floorStart;

			if order.Floor >= floorLower && order.Floor <= floorUpper {
				return true;
			}
		}
	}

	return false;
}

func GetIndexInQueue(order Order, elevatorLastFloorVisited int, isElevatorMoving bool, elevatorDirection Direction) int {

	// Empty list
	if len(orders) < 1 {
		return 0;

	//-----------------------------------------------//

	} else {

		// Check if we should set it in first
		startFloor := elevatorLastFloorVisited;

		if isElevatorMoving { // If we have left the elevatorLastFloorVisited

			if elevatorDirection == DIRECTION_UP && startFloor < config.NUMBER_OF_FLOORS - 1 {
				
				startFloor = startFloor + 1;

			} else if elevatorDirection == DIRECTION_DOWN && startFloor > 0 {

				startFloor = startFloor - 1;
			}
		}

		if shouldOrderBeBetween(order, startFloor, orders[0].Floor) {
			return 0;
		}

		// Check if it should be in between any orders currently taken
		for orderIndex := range orders {

			if orderIndex >= 0 && (orderIndex + 1) < len(orders) {

				floorStart := orders[orderIndex].Floor;
				floorEnd   := orders[orderIndex + 1].Floor;

				if shouldOrderBeBetween(order, floorStart, floorEnd) {
					return orderIndex + 1;
				}
			}	
		}

		// Not found, thus it must be last
		return len(orders);
	}
}

//-----------------------------------------------//

func Add(order Order, elevatorLastFloorVisited int, isElevatorMoving bool, elevatorDirection Direction) {

	index := GetIndexInQueue(order, elevatorLastFloorVisited, isElevatorMoving, elevatorDirection);
	
	if index == 0 {

		orders = append([]Order{ order }, orders ... );
		
	} else if index >= len(orders) {
		
		orders = append(orders, order);

	} else {
		
		orders = append(orders, order);

		// Bubble order down
		for orderIndex := (len(orders) - 1); orderIndex > index; orderIndex-- {
			
			tempOrder := orders[orderIndex];

			orders[orderIndex] 		= orders[orderIndex - 1];
			orders[orderIndex - 1] 	= tempOrder;
		}
	}
}

//-----------------------------------------------//

func GetCostOf(order Order, elevatorLastFloorVisited int, isElevatorMoving bool, elevatorDirection Direction) int {
	
	index := GetIndexInQueue(order, elevatorLastFloorVisited, isElevatorMoving, elevatorDirection);

	if index == 0 {

		return int(math.Abs(float64(order.Floor - elevatorLastFloorVisited)));
	
	} else {

		cost := 0.0;

		// First move
		if orders[0].Floor != elevatorLastFloorVisited {
			cost = math.Abs(float64(orders[0].Floor - elevatorLastFloorVisited));
			cost = cost + COST_FOR_STOP;
		}

		// Between orders
		for orderIndex := 1; orderIndex <= index - 1; orderIndex++ {

			if orders[orderIndex].Floor != orders[orderIndex - 1].Floor {
				cost = cost + math.Abs(float64(orders[orderIndex].Floor - orders[orderIndex - 1].Floor)); 	// Distance to travel
				cost = cost + COST_FOR_STOP;
			}
		}

		// Last step
		cost = cost + math.Abs(float64(order.Floor - orders[index - 1].Floor));

		return int(cost);
	}
}