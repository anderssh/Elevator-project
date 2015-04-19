package ordersGlobal;

import(
	. "user/typeDefinitions"
	"user/log"
);

//-----------------------------------------------//

var orders []OrderGlobal = make([]OrderGlobal, 0, 1);

//-----------------------------------------------//

func MakeFromOrder(order Order, responsibleWorkerIPAddr string) OrderGlobal {
	return OrderGlobal{ ResponsibleWorkerIPAddr : responsibleWorkerIPAddr, Type : order.Type, Floor : order.Floor };
}

//-----------------------------------------------//

func AlreadyStored(order Order) bool {
	
	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type  && orders[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

func Add(order OrderGlobal) {
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

//-----------------------------------------------//

func GetAll() []OrderGlobal {
	return orders;
}

//-----------------------------------------------//

func MergeWith(ordersToMerge []OrderGlobal) {

	log.Data(orders)
	log.Data(ordersToMerge)
}

//-----------------------------------------------//

func Display() {

	log.DataWithColor(log.COLOR_BLUE, "Global orders:");

	for orderIndex := range orders {
		log.DataWithColor(log.COLOR_BLUE, orders[orderIndex]);
	}
}