package orders;

import(
	. "../typeDefinitions"
	"../log"
);

//-----------------------------------------------//

var orders []Order;

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

				if (orderIndex == len(orders) - 1) {
					orders = orders[:(len(orders) - 1)];
				} else {
					orders = append(orders[0:orderIndex], orders[orderIndex + 1:] ... );
				}

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

	// In between is all that is need, not dependent on moving direction
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

			if elevatorDirection == DIRECTION_UP && startFloor < 4 {
				
				startFloor = startFloor + 1;

			} else if elevatorDirection == DIRECTION_DOWN && startFloor > 1 {

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
	
	if index >= len(orders) {

		log.Data("Add order last");
		orders = append(orders, order);
		
	} else if index == 0 {
		
		log.Data("Add order first");
		orders = append([]Order{ order }, orders ... );

	} else {
		
		log.Data("Add order in the middle somewhere at ", index);
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

func OrderFromButtonFloor(button ButtonFloor) Order {
	
	if (button.Type == BUTTON_CALL_UP) {
		return Order{ Type : ORDER_UP, Floor : button.Floor };
	} else if (button.Type == BUTTON_CALL_DOWN) {
		return Order{ Type : ORDER_DOWN, Floor : button.Floor };
	} else {
		return Order{ Type : ORDER_INSIDE, Floor : button.Floor };
	}
}