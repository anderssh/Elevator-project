package elevatorController;

import(
	. "../typeDefinitions"
	"../network"
	"../encoder/JSON"
);

//-----------------------------------------------//

var isCurrentlyMaster bool = true;

func master(broadcastChannel chan network.Message, addServerRecipientChannel chan network.Recipient) {

	newOrderRecipient := network.Recipient{ Name : "masterNewOrder", Channel : make(chan []byte) };
	//newOrderRecipient := network.Recipient{ Name : "masterNewOrder", Channel : make(chan []byte) };

	addServerRecipientChannel <- newOrderRecipient;
	//addServerRecipientChannel <- newOrderRecipient;
	
	for {
		select {
			case orderEncoded := <- newOrderRecipient.Channel:
				
				if isCurrentlyMaster {

					var order Order;
					err := JSON.Decode(orderEncoded, &order);

					if err != nil {

					}

					broadcastChannel <- network.Message{ RecipientName : "receiveNewDestinationOrder", Data : orderEncoded };

					// <- newOrder;

					// If not received
					// Ask all slaves
					// Wait for response cost for some time
					// Send order to best slave
						// Wait for ack

				} else {
					// Dont care
				}
		}	
	}
}
