package ordersGlobal;

import(
	. "../typeDefinitions"
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