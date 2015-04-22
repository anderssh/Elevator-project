package network;

import(
	"net"
	"strconv"
	"sync"
	"user/config"
	"user/log"
	"user/encoder/JSON"
	"strings"
	"time"
);

//-----------------------------------------------//

var tcpConnections 			map[string]*net.TCPConn;
var tcpConnectionsMutex 	*sync.Mutex;

func deleteConnectionWithIPAddr(iPAddrToDelete string) {
	
	tcpConnectionsMutex.Lock();
	
	for remoteAddr, connection := range tcpConnections {

		if strings.HasPrefix(remoteAddr, iPAddrToDelete + ":") {
			connection.Close();
			delete(tcpConnections, remoteAddr);
		}
	}
	
	tcpConnectionsMutex.Unlock();
}

//-----------------------------------------------//

func tcpListenOnConnection(listenConnection *net.TCPConn, remoteAddr *net.TCPAddr, remoteIPAddr string, messageChannel chan<- Message, eventDisconnect chan string) {

	messageBuffer := make([]byte, 4096);
	log.Warning("Start", remoteAddr, remoteIPAddr)
	for {

		if remoteIPAddr != localIPAddr && remoteIPAddr != LOCALHOST {  									// For timeout checking
			listenConnection.SetReadDeadline(time.Now().Add(config.TCP_READ_CONNECTION_DEADLINE));
		}
		
		messageLength, err := listenConnection.Read(messageBuffer);
	
		if err != nil || messageLength < 0 {

			log.Error("Network: Error when reading from TCP.", remoteAddr.String(), remoteIPAddr);

			deleteConnectionWithIPAddr(remoteIPAddr);

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

	connection, err := net.DialTCP("tcp", nil, remoteAddr);

	if err != nil {
		
		log.Error("Network: Could not dial tcp", err, remoteIPAddr, remoteAddr);

		eventDisconnect <- remoteIPAddr;
		
		return;

	} else {

		tcpConnectionsMutex.Lock();
		tcpConnections[remoteAddr.String()] = connection;
		tcpConnectionsMutex.Unlock();

		return;
	}
}

func TCPTransmitServer(transmitChannel chan Message, eventDisconnect chan string) {

	for {
		select {
			case message := <- transmitChannel:

				remoteAddr, err := net.ResolveTCPAddr("tcp", message.DestinationIPAddr + ":" + strconv.Itoa(config.PORT_SERVER_DEFAULT));

				if err != nil {
					log.Error(err);
				}
				
				_, connectionExists := tcpConnections[remoteAddr.String()];

				if !connectionExists {
					tcpConnectTo(remoteAddr, message.DestinationIPAddr, eventDisconnect);
				}

				tcpConnectionsMutex.Lock();

				sendConnection, connectionExists := tcpConnections[remoteAddr.String()];

				if !connectionExists {

					tcpConnectionsMutex.Unlock();
					log.Error("Failed to add connection to list", remoteAddr.String());

				} else {

					encodedMessage, _ 	:= JSON.Encode(message);
					n, err 			  	:= sendConnection.Write(encodedMessage);

					tcpConnectionsMutex.Unlock();

					if err != nil || n < 0 {
						deleteConnectionWithIPAddr(message.DestinationIPAddr);
					}
				}
		}
	}
}