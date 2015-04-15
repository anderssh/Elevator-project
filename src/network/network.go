package network;

import(
	"net"
	"strconv"
	"time"
	"../log"
	"../encoder/JSON"
);

//-----------------------------------------------//

const (
	BROADCAST_ADDR 	string = "255.255.255.255"
	LOCALHOST 		string = "localhost"

	PORT_SERVER_DEFAULT 		int = 9125
	PORT_SERVER_WITH_TIMEOUT	int = 9126
);

//-----------------------------------------------//

var iPAddr 	string;

func GetLocalIPAddr() string {
	return iPAddr;
}

//-----------------------------------------------//

func Initialize(){

    adresses, err := net.InterfaceAddrs();
    if err != nil {
        log.Error("Error in finding all Interface adresses");
    }

    for _, address := range adresses {

        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() { 		// check the address type and if it is not a loopback the display it
            if ipnet.IP.To4() != nil {
              iPAddr = ipnet.IP.String();
            }
        }
	}
}

//-----------------------------------------------//

type Message struct {
	RecipientID 		string;
	
	DestinationIPAddr 	string;
	DestinationPort 	int;
	
	SenderIPAddr		string;
	SenderPort			int;
	
	Data 				[]byte;
}

func MakeMessage(recipientID string, data []byte, destinationIPAddr string) Message {
	
	return Message{	RecipientID : recipientID, 
					
					DestinationIPAddr : destinationIPAddr, 
					DestinationPort : PORT_SERVER_DEFAULT,
					
					SenderIPAddr : iPAddr,
					SenderPort : PORT_SERVER_DEFAULT,
	 				
	 				Data : data }
}

func MakeTimeoutMessage(recipientID string, data []byte, destinationIPAddr string) Message {
	
	return Message{	RecipientID : recipientID, 
					
					DestinationIPAddr : destinationIPAddr, 
					DestinationPort : PORT_SERVER_WITH_TIMEOUT,
					
					SenderIPAddr : iPAddr,
					SenderPort : PORT_SERVER_WITH_TIMEOUT,
	 				
	 				Data : data }
}

type Recipient struct {
	ID 				string;
	ReceiveChannel 	chan Message;
}

//-----------------------------------------------//

func listen(IPAddr string, messageChannel chan<- Message) {

	listenAddress, _     := net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(PORT_SERVER_DEFAULT));
	listenConnection, err := net.ListenUDP("udp", listenAddress);

	if err != nil{
		log.Error(err)
	}
	
	defer func() {
		if errRecovered := recover(); errRecovered != nil {
			listenConnection.Close();
		}
	}();

	messageBuffer := make([]byte, 1024);

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

func ListenServer(IPAddr string, addRecipientChannel chan Recipient) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go listen(IPAddr, messageChannel);

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

func listenWithTimeout(IPAddr string, messageChannel chan<- Message, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	listenAddress, _ 	:= net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(PORT_SERVER_WITH_TIMEOUT));
	log.Error(listenAddress)
	listenConnection, err := net.ListenUDP("udp", listenAddress);
	log.Error(err)
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

	messageBuffer := make([]byte, 1024);

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

func ListenServerWithTimeout(IPAddr string, addRecipientChannel chan Recipient, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go listenWithTimeout(IPAddr, messageChannel, deadlineDuration, timeoutNotifier);

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

func TransmitServer(transmitChannel chan Message) {

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

//-----------------------------------------------//

func Repeat(transmitChannel chan Message, message Message, repeatCount int, delayInMilliseconds int64){

	for i := 0; i < repeatCount; i++ {
		transmitChannel <- message;
		time.Sleep(time.Duration(delayInMilliseconds) *time.Millisecond);
	}

}