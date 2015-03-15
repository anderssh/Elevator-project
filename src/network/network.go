package network;

import(
	"net"
	"strconv"
	"time"
	"../log"
);

//-----------------------------------------------//

type NetworkMessage struct {
	Length 			int;
	Data 			string;
	RemoteAddress 	*net.UDPAddr;
}

//-----------------------------------------------//

func GetNewAddress(IPAddress string, port int) *net.UDPAddr {
	addr, _ := net.ResolveUDPAddr("udp", IPAddress + ":" + strconv.Itoa(port));

	return addr;
}

//-----------------------------------------------//

func Listen(listenAddress *net.UDPAddr, listenChannel chan string) {

	

}

func ListenWithDeadline(listenAddress *net.UDPAddr, listenChannel chan string, deadlineDuration time.Duration) error {
	
	listenConnection, _ := net.ListenUDP("udp", listenAddress);
	listenConnection.SetDeadline(time.Now().Add(deadlineDuration));

	messageBuffer := make([]byte, 1024);

	for {
		messageLength, _, err := listenConnection.ReadFromUDP(messageBuffer);

		if err != nil {

			log.Error(err);
			return err;

		} else {

			listenConnection.SetDeadline(time.Now().Add(deadlineDuration));
			listenChannel <- string(messageBuffer[0:messageLength]);
		}
	}
}

//-----------------------------------------------//

func Send(sendAddress *net.UDPAddr, sendChannel chan string) {
	
	for {
		select {
			case message := <- sendChannel:

				sendConnection, _ := net.DialUDP("udp", nil, sendAddress);
				sendConnection.Write([]byte(message));
		}
	}
}

//----------------------------------------


/*
func listen(conn *net.UDPConn) {
buffer := make([]byte, 1024);
for {
messageSize, _, _ := conn.ReadFromUDP(buffer);
fmt.Println("listend: " + string(buffer[0:messageSize]));
}
}
func transmit(conn *net.UDPConn) {
for {
time.Sleep(2000*time.Millisecond);
message := "Hello server";
conn.Write([]byte(message));
fmt.Println("Sent: " + message);
}
}
func main() {
serverIP := "129.241.187.255";
serverPort := 20016;
serverAddr, _ := net.ResolveUDPAddr("udp", serverIP + ":" + strconv.Itoa(serverPort));
listenPort := 20016;
listenAddr, _ := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(listenPort));
fmt.Println(listenAddr);
fmt.Println(serverAddr);
listenConn, _ := net.ListenUDP("udp", listenAddr);
transmitConn, _ := net.DialUDP("udp", nil, serverAddr);
go listen(listenConn);
go transmit(transmitConn);
d_chan := make(chan bool, 1);
<- d_chan;
*/