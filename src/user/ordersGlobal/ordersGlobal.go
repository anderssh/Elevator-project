package ordersGlobal;

import(
	. "user/typeDefinitions"
);

//-----------------------------------------------//

type OrderGlobal struct {
	IPAddr 	string;
	order 	Order;
}

var orders []Order = make([]Order, 0, 1);

//-----------------------------------------------//

func AlreadyStored(order Order) bool {
	
	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type  && orders[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

func Add(order Order) {
	orders = append(orders, order);
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