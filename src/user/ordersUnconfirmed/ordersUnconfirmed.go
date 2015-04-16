package ordersUnconfirmed;

import(
	. "user/typeDefinitions"
);

//-----------------------------------------------//

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

func Remove(order Order) {

	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type && orders[orderIndex].Floor == order.Floor {
			
			if (orderIndex == len(orders) - 1) {
				orders = orders[:(len(orders) - 1)];
			} else {
				orders = append(orders[0:orderIndex], orders[orderIndex + 1:] ... );
			}
		}
	}
}