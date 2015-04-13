package network;

import(
	"net"
	"strconv"
	"time"
	"../log"
	"../encoder/JSON"
);

//-----------------------------------------------//

type Message struct {
	RecipientName 		string;
	DestinationIPAddr 	string;
	DestinationPort 	int;
	
	SenderIP			string;
	SenderPort			int;
	Data 				[]byte;
}


type Recipient struct {
	Name 		string;
	ChannelMessage 	chan Message;
}

//-----------------------------------------------//

var IPAddress 	string;
var port 		int;
func initialize(){


    adresses, err := net.InterfaceAddrs()

    if err != nil {
        log.Error("Error in finding all Interface adresses")
    }

    for _, address := range adresses {

          // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
              IPAddress = ipnet.IP.String();
               }
         }
	}
}

//-----------------------------------------------//

func MakeMessage (recipientName string, data []byte, destinationIPAddr string, destinationPort int) Message {
	return 	Message{RecipientName : recipientName, 
			DestinationIPAddr : destinationIPAddr, 
			DestinationPort : destinationPort,
			SenderIP : IPAddress,
			SenderPort : port,
	 		Data : data}
}

func listen(IPAddr string, port int, messageChannel chan<- Message) {

	listenAddress, _     := net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(port));
	listenConnection, _ := net.ListenUDP("udp", listenAddress);

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

func ListenServer(IPAddr string, port int, addRecipientChannel chan Recipient) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go listen(IPAddr, port, messageChannel);

	for {
		select {
			case message := <- messageChannel:
				
				for recipientIndex := range recipients {
					if message.RecipientName == recipients[recipientIndex].Name {
						
						recipients[recipientIndex].ChannelMessage <- message;
						break;
					}
				}

			case newRecipient := <- addRecipientChannel:
				
				recipients = append(recipients, newRecipient);
		}
	}
}

//-----------------------------------------------//

func listenWithTimeout(IPAddr string, port int, messageChannel chan<- Message, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	listenAddress, _ 	:= net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(port));
	listenConnection, _ := net.ListenUDP("udp", listenAddress);
	
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

func ListenServerWithTimeout(IPAddr string, port int, addRecipientChannel chan Recipient, deadlineDuration time.Duration, timeoutNotifier chan<- bool) {

	recipients 		:= make([]Recipient, 1);
	messageChannel 	:= make(chan Message);

	go listenWithTimeout(IPAddr, port, messageChannel, deadlineDuration, timeoutNotifier);

	for {
		select {
			case message := <- messageChannel:
				
				for recipientIndex := range recipients {
					if message.RecipientName == recipients[recipientIndex].Name {
						recipients[recipientIndex].ChannelMessage <- message;
						break;
					}
				}

			case newRecipient := <- addRecipientChannel:
				
				recipients = append(recipients, newRecipient);
		}
	}
}

//-----------------------------------------------//

func TransmitServer(IPAddr string, port int, sendChannel chan Message) {
	
	transmitAddr, _ := net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(port));

	for {
		select {
			case message := <- sendChannel:

				encodedMessage, _ := JSON.Encode(message);

				sendConnection, _ := net.DialUDP("udp", nil, transmitAddr);
				sendConnection.Write(encodedMessage);
		}
	}
}