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

func AlreadyStored(order OrderLocal) bool {
	
	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type  && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			return true;
		}
	}

	return false;
}

//-----------------------------------------------//

func Add(order OrderLocal, eventUnconfirmedOrderTimeout chan OrderLocal) {

	timer := time.AfterFunc(config.TIMEOUT_TIME_ORDER_TAKEN, func() {
		eventUnconfirmedOrderTimeout <- order;
	});

	ordersUnconfirmed = append(ordersUnconfirmed, OrderUnconfirmed{ Timer : timer, Type : order.Type, Floor : order.Floor });
}

//-----------------------------------------------//


func Remove(order OrderLocal) {

	for orderIndex := range ordersUnconfirmed {
		if ordersUnconfirmed[orderIndex].Type == order.Type && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			
			ordersUnconfirmed[orderIndex].Timer.Stop();
			ordersUnconfirmed = append(ordersUnconfirmed[0:orderIndex], ordersUnconfirmed[orderIndex + 1:] ... );
			return;
		}
	}
}

func ResetTimer(order OrderLocal, eventUnconfirmedOrderTimeout chan OrderLocal) {

	for orderIndex := range ordersUnconfirmed {
		
		if ordersUnconfirmed[orderIndex].Type == order.Type  && ordersUnconfirmed[orderIndex].Floor == order.Floor {
			
			ordersUnconfirmed[orderIndex].Timer.Stop();

			timer := time.AfterFunc(config.TIMEOUT_TIME_ORDER_TAKEN, func() {
				eventUnconfirmedOrderTimeout <- order;
			});

			ordersUnconfirmed[orderIndex].Timer = timer;

			return;
		}
	}

	log.Error("The order to be reset is not in ordersUnconfirmed")
}