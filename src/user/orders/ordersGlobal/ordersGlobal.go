package ordersGlobal;

import(
	. "user/typeDefinitions"
	"user/log"
	"time"
);

//-----------------------------------------------//

var orders []OrderGlobal = make([]OrderGlobal, 0, 1);

//-----------------------------------------------//

func GetAll() []OrderGlobal {
	return orders;
}

func SetTo(newOrders []OrderGlobal) {
	orders = newOrders;
}

//-----------------------------------------------//

func MakeFromOrder(order Order, responsibleWorkerIPAddr string) OrderGlobal {
	return OrderGlobal{ ResponsibleWorkerIPAddr : responsibleWorkerIPAddr, Type : order.Type, Floor : order.Floor };
}

func MakeBackup() OrdersGlobalBackup {
	return OrdersGlobalBackup{ Orders : orders, Timestamp : time.Now().UnixNano() };
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

func UpdateResponsibility(order OrderGlobal) {

	for orderIndex := range orders {
		if orders[orderIndex].Type == order.Type  && orders[orderIndex].Floor == order.Floor {
			orders[orderIndex].ResponsibleWorkerIPAddr = order.ResponsibleWorkerIPAddr;
			return;
		}
	}
}

func ResetResponsibilityOnWorkerIPAddr(workerIPAddr string) {

	for orderIndex := range orders {
		if orders[orderIndex].ResponsibleWorkerIPAddr == workerIPAddr {
			orders[orderIndex].ResponsibleWorkerIPAddr = "";
		}
	}
}

func ResetAllResponsibilities() {

	for orderIndex := range orders {
		orders[orderIndex].ResponsibleWorkerIPAddr = "";
	}
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

func HasOrderToRedistribute() bool {

	for orderIndex := range orders {
		if orders[orderIndex].ResponsibleWorkerIPAddr == "" {
			return true;
		}
	} 

	return false;
}

func GetOrderToRedistribute() OrderGlobal {

	for orderIndex := range orders {
		if orders[orderIndex].ResponsibleWorkerIPAddr == "" {
			return orders[orderIndex];
		}
	}

	return OrderGlobal{}; // Fix this
}

//-----------------------------------------------//

func MergeWith(ordersToMerge []OrderGlobal) {

	// Add all orders not currently in list
	for orderToMergeIndex := range ordersToMerge {

		orderToMerge := ordersToMerge[orderToMergeIndex];
		orderToMergeAllreadyStored := false;

		for orderIndex := range orders {

			order := orders[orderIndex];

			if order.Type == orderToMerge.Type && order.Floor == orderToMerge.Floor {
				orderToMergeAllreadyStored = true;
				continue;
			}
		}

		if !orderToMergeAllreadyStored {
			orders = append(orders, orderToMerge);
		}
	}
}

//-----------------------------------------------//

func Display() {

	log.DataWithColor(log.COLOR_BLUE, "Global orders:");

	for orderIndex := range orders {
		log.DataWithColor(log.COLOR_BLUE, orders[orderIndex]);
	}
}