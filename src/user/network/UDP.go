package network;

import(
	"net"
	"strconv"
	"time"
	"user/config"
	"user/log"
	"user/encoder/JSON"
);

//-----------------------------------------------//

func udpListen(IPAddr string, messageChannel chan<- Message) {

	listenAddress, _     := net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(config.PORT_SERVER_DEFAULT));
	listenConnection, err := net.ListenUDP("udp", listenAddress);

	if err != nil{
		log.Error(err);
	}
	
	defer func() {
		if errRecovered := recover(); errRecovered != nil {
			listenConnection.Close();
		}
	}();

	messageBuffer := make([]byte, 4096);

	for {
		messageLength, _, err := listenConnection.ReadFromUDP(messageBuffer);
	
		if err != nil {
			
			panic(err);

		} else {

			var decodedMessage Message;
			originalMessage := messageBuffer[0:messageLength];
			JSON.Decode(originalMessage, &decodedMessage);

			messageChannel <- decodedMessage;
		}
	}
}

func UDPListenServer(IPAddr string, addRecipientChannel chan Recipient) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go udpListen(IPAddr, messageChannel);

	for {
		select {
			case message := <- messageChannel:
				
				for recipientIndex := range recipients {
					if message.RecipientID == recipients[recipientIndex].ID {
						
						recipients[recipientIndex].ReceiveChannel <- message;
						break;
					}
				}

			case newRecipient := <- addRecipientChannel:
				
				recipients = append(recipients, newRecipient);
		}
	}
}

//-----------------------------------------------//

func udpListenWithTimeout(IPAddr string, messageChannel chan<- Message, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	listenAddress, _ 	:= net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(config.PORT_SERVER_WITH_TIMEOUT));
	listenConnection, err := net.ListenUDP("udp", listenAddress);

	if err != nil {
		log.Error(err);
	}
	
	listenConnection.SetDeadline(time.Now().Add(deadlineDuration));

	defer func() {
		if errRecovered := recover(); errRecovered != nil {
			
			if errNet, ok := errRecovered.(net.Error); ok && errNet.Timeout() {
				log.Warning("Listen server with deadline timed out");
			} else {
				log.Error("Unknown listen server timeout");
			}

			listenConnection.Close();
			timeoutNotifier <- true;
		}
	}();

	messageBuffer := make([]byte, 4096);

	for {
		messageLength, _, err := listenConnection.ReadFromUDP(messageBuffer);
	
		if err != nil {

			panic(err);

		} else {

			listenConnection.SetDeadline(time.Now().Add(deadlineDuration));

			var decodedMessage Message;
			originalMessage := messageBuffer[0:messageLength];
			JSON.Decode(originalMessage, &decodedMessage);

			messageChannel <- decodedMessage;
		}
	}
}

func UDPListenServerWithTimeout(IPAddr string, addRecipientChannel chan Recipient, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go udpListenWithTimeout(IPAddr, messageChannel, deadlineDuration, timeoutNotifier);

	for {
		select {
			case message := <- messageChannel:
				
				for recipientIndex := range recipients {
					if message.RecipientID == recipients[recipientIndex].ID {
						recipients[recipientIndex].ReceiveChannel <- message;
						break;
					}
				}

			case newRecipient := <- addRecipientChannel:
				
				recipients = append(recipients, newRecipient);
		}
	}
}

//-----------------------------------------------//

func UDPTransmitServer(transmitChannel chan Message) {

	for {
		select {
			case message := <- transmitChannel:

				transmitAddr, _   := net.ResolveUDPAddr("udp", message.DestinationIPAddr + ":" + strconv.Itoa(message.DestinationPort));
				encodedMessage, _ := JSON.Encode(message);

				sendConnection, _ := net.DialUDP("udp", nil, transmitAddr);
				sendConnection.Write(encodedMessage);
		}
	}
}