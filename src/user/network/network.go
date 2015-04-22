package network;

import(
	"net"
	"time"
	"sync"
	"user/config"
);

//-----------------------------------------------//

const (
	BROADCAST_ADDR 	string = "255.255.255.255"
	LOCALHOST 		string = "127.0.0.1"
);

//-----------------------------------------------//

var localIPAddr 	string;

func GetLocalIPAddr() string {
	return localIPAddr;
}

//-----------------------------------------------//

func Initialize(){

    discoverAddr, _ := net.ResolveUDPAddr("udp", BROADCAST_ADDR + ":50000");
    discoverConn, _ := net.DialUDP("udp", nil, discoverAddr);
	
	discoverConnAddr := discoverConn.LocalAddr();
	localAddr, _ := net.ResolveUDPAddr("udp", discoverConnAddr.String());
	
	localIPAddr = localAddr.IP.String();
	
	discoverConn.Close();

	tcpConnections 		= make(map[string]*net.TCPConn);
	tcpConnectionsMutex = &sync.Mutex{};
}

//-----------------------------------------------//

type Recipient struct {
	ID 				string;
	ReceiveChannel 	chan Message;
}

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
					
					SenderIPAddr : localIPAddr,
					SenderPort : config.PORT_SERVER_DEFAULT,
	 				
	 				Data : data }
}

func MakeTimeoutServerMessage(recipientID string, data []byte, destinationIPAddr string) Message {
	
	return Message{	RecipientID : recipientID, 
					
					DestinationIPAddr : destinationIPAddr, 
					DestinationPort : config.PORT_SERVER_WITH_TIMEOUT,
					
					SenderIPAddr : localIPAddr,
					SenderPort : config.PORT_SERVER_WITH_TIMEOUT,
	 				
	 				Data : data }
}

//-----------------------------------------------//

func Repeat(transmitChannel chan Message, message Message, repeatCount int, delayInMilliseconds int64){

	for i := 0; i < repeatCount; i++ {
		transmitChannel <- message;
		time.Sleep(time.Duration(delayInMilliseconds) * time.Millisecond);
	}
}