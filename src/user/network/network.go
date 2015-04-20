package network;

import(
	"net"
	"strconv"
	"time"
	"sync"
	"user/config"
	"user/log"
	"user/encoder/JSON"
);

//-----------------------------------------------//

const (
	BROADCAST_ADDR 	string = "255.255.255.255"
	LOCALHOST 		string = "localhost"
);

//-----------------------------------------------//

var iPAddr 	string;

func GetLocalIPAddr() string {
	return iPAddr;
}

//-----------------------------------------------//

var tcpConnections 			map[string]*net.TCPConn;
var tcpConnectionsMutex 	*sync.Mutex;

func Initialize(){

    discoverAddr, _ := net.ResolveUDPAddr("udp", BROADCAST_ADDR + ":50000");
    discoverConn, _ := net.DialUDP("udp4", nil, discoverAddr);
	
	discoverConnAddr := discoverConn.LocalAddr();
	localAddr, _ := net.ResolveUDPAddr("udp4", discoverConnAddr.String());
	
	iPAddr = localAddr.IP.String();
	
	discoverConn.Close();

	tcpConnections 		= make(map[string]*net.TCPConn);
	tcpConnectionsMutex = &sync.Mutex{};
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
					DestinationPort : config.PORT_SERVER_DEFAULT,
					
					SenderIPAddr : iPAddr,
					SenderPort : config.PORT_SERVER_DEFAULT,
	 				
	 				Data : data }
}

func MakeTimeoutMessage(recipientID string, data []byte, destinationIPAddr string) Message {
	
	return Message{	RecipientID : recipientID, 
					
					DestinationIPAddr : destinationIPAddr, 
					DestinationPort : config.PORT_SERVER_WITH_TIMEOUT,
					
					SenderIPAddr : iPAddr,
					SenderPort : config.PORT_SERVER_WITH_TIMEOUT,
	 				
	 				Data : data }
}

type Recipient struct {
	ID 				string;
	ReceiveChannel 	chan Message;
}

//-----------------------------------------------//

func udpListen(IPAddr string, messageChannel chan<- Message) {

	listenAddress, _     := net.ResolveUDPAddr("udp", IPAddr + ":" + strconv.Itoa(config.PORT_SERVER_DEFAULT));
	listenConnection, err := net.ListenUDP("udp", listenAddress);

	if err != nil{
		log.Error(err)
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

//-----------------------------------------------//

func tcpListenOnConnection(listenConnection *net.TCPConn, remoteAddr *net.TCPAddr, remoteIPAddr string, messageChannel chan<- Message, eventDisconnect chan string) {

	messageBuffer := make([]byte, 4096);

	for {
		messageLength, err := listenConnection.Read(messageBuffer);
	
		if err != nil || messageLength < 0 {

			log.Error("Network: Error when reading from TCP.");

			tcpConnectionsMutex.Lock();
			listenConnection.Close();
			delete(tcpConnections, remoteAddr.String());
			tcpConnectionsMutex.Unlock();

			eventDisconnect <- remoteIPAddr;

			return;

		} else {

			var decodedMessage Message;
			originalMessage := messageBuffer[0:messageLength];
			JSON.Decode(originalMessage, &decodedMessage);

			messageChannel <- decodedMessage;
		}
	}
}

func tcpListen(IPAddr string, messageChannel chan<- Message, eventDisconnect chan string) {

	serverAddr, _     		:= net.ResolveTCPAddr("tcp", IPAddr + ":" + strconv.Itoa(config.PORT_SERVER_DEFAULT));
	serverConnection, err 	:= net.ListenTCP("tcp", serverAddr);
	
	if err != nil{
		log.Error(err)
	}

	for {

		log.DataWithColor(log.COLOR_GREEN, "Network: Waiting for new connect");

		listenConnection, _ := serverConnection.AcceptTCP();
		remoteAddrRaw 		:= listenConnection.RemoteAddr();
		remoteAddr, _ 		:= net.ResolveTCPAddr("tcp", remoteAddrRaw.String());
		remoteIPAddr 		:= remoteAddr.IP.String();

		log.DataWithColor(log.COLOR_GREEN, "Network: Connected to", remoteIPAddr);

		tcpConnectionsMutex.Lock();
		tcpConnections[remoteAddr.String()] = listenConnection;
		tcpConnectionsMutex.Unlock();

		go tcpListenOnConnection(listenConnection, remoteAddr, remoteIPAddr, messageChannel, eventDisconnect);
	}
}

func TCPListenServer(IPAddr string, addRecipientChannel chan Recipient, eventDisconnect chan string) {

	recipients 		:= make([]Recipient, 0, 1);
	messageChannel 	:= make(chan Message);

	go tcpListen(IPAddr, messageChannel, eventDisconnect);

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

func tcpConnectTo(remoteAddr *net.TCPAddr, remoteIPAddr string, eventDisconnect chan string) {

	for {

		connection, err := net.DialTCP("tcp", nil, remoteAddr);

		if err != nil {
			
			log.Error("Network: Could not dial tcp", remoteIPAddr, remoteAddr);
			log.Error(err)

			eventDisconnect <- remoteIPAddr;
			
			return;

		} else {

			tcpConnectionsMutex.Lock();
			tcpConnections[remoteAddr.String()] = connection;
			tcpConnectionsMutex.Unlock();

			return;
		}
	}
}

func TCPTransmitServer(transmitChannel chan Message, eventDisconnect chan string) {

	for {
		select {
			case message := <- transmitChannel:

				remoteAddr, err 	:= net.ResolveTCPAddr("tcp", message.DestinationIPAddr + ":" + strconv.Itoa(config.PORT_SERVER_DEFAULT));

				if err != nil {
					log.Error(err);
				}

				_, connectionExists := tcpConnections[remoteAddr.String()];

				if !connectionExists {
					tcpConnectTo(remoteAddr, message.DestinationIPAddr, eventDisconnect);
				}

				tcpConnectionsMutex.Lock();

				sendConnection, _ := tcpConnections[remoteAddr.String()];
				encodedMessage, _ := JSON.Encode(message);
				n, err 			  := sendConnection.Write(encodedMessage);

				tcpConnectionsMutex.Unlock();

				if err != nil || n < 0 {
					tcpConnectionsMutex.Lock();
					sendConnection.Close();
					delete(tcpConnections, remoteAddr.String());
					tcpConnectionsMutex.Unlock();
				}
		}
	}
}

//-----------------------------------------------//

func Repeat(transmitChannel chan Message, message Message, repeatCount int, delayInMilliseconds int64){

	for i := 0; i < repeatCount; i++ {
		transmitChannel <- message;
		time.Sleep(time.Duration(delayInMilliseconds) * time.Millisecond);
	}
}