package ordersGlobal;

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

func AlreadyStored(order Order) bool {
	
	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type  && orders[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

func Add(order Order) {
	orders = append(orders, order[]);
}