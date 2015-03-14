// Store orders locally in the order they are supposed to be taken

package orders;

import(
	. "../typeDefinitions"
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

func AllreadyStored(order Order) bool {
	
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

func RemoveTop() {
	orders = orders[1:];
}

func Add(order Order) {
	orders = append(orders, order);
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