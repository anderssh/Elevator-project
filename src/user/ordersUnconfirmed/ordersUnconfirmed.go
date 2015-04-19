package ordersUnconfirmed;

import(
	. "user/typeDefinitions"
	"time"
	"user/config"
	"user/log"
);

//-----------------------------------------------//

var ordersUnconfirmed []OrderUnconfirmed = make([]OrderUnconfirmed, 0, 1);

//-----------------------------------------------//

func AlreadyStored(order Order) bool {
	
	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type  && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

func Add(order Order, eventUnconfirmedOrderTimeout chan Order) {

	timer := time.AfterFunc(config.TIMEOUT_TIME_ORDER_TAKEN, func() {
		eventUnconfirmedOrderTimeout <- order;
		log.Warning("Unconfirmed order timeout")
	});
	log.Warning("Legger til ", len(ordersUnconfirmed), order)
	ordersUnconfirmed = append(ordersUnconfirmed, OrderUnconfirmed{Type : order.Type, Floor : order.Floor, Timer : timer})
	log.Warning("Lagt til ", len(ordersUnconfirmed), ordersUnconfirmed)
}
//-----------------------------------------------//

func Remove(order Order) {

	log.Warning("Remove order")

	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			
			ordersUnconfirmed[orderIndex].Timer.Stop();
			ordersUnconfirmed = append(ordersUnconfirmed[0:orderIndex], ordersUnconfirmed[orderIndex + 1:] ... );
			return;
		}
	}
}

func ResetTimer(order Order) {

	log.Warning("I ResetTimer");
	log.Error("Lengden av lista er :", len(ordersUnconfirmed), ordersUnconfirmed);

	for orderIndex := range ordersUnconfirmed {
		
		if ordersUnconfirmed[orderIndex].Type == order.Type  && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			
			ordersUnconfirmed[orderIndex].Timer.Reset(config.TIMEOUT_TIME_ORDER_TAKEN);
			return;
		}	
	}

	log.Error("The order to be resent is not in ordersUnconfirmed")
}