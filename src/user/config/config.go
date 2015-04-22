package config;

import(
	"time"
);

const(
	// Network

	PORT_SERVER_DEFAULT 								int 			= 	9125
	PORT_SERVER_WITH_TIMEOUT							int 			= 	9126

	TCP_READ_CONNECTION_DEADLINE 						time.Duration 	= 	time.Millisecond * 200

	// Process pairs

	SHOULD_USE_PROCESS_PAIRS 							bool 			= 	false

	BACKUP_PROCESS_ALIVE_MESSAGE_DEADLINE				time.Duration	=   time.Millisecond * 200
	BACKUP_PROCESS_ALIVE_NOTIFICATION_SLEEP  			time.Duration	=   time.Millisecond * 15

	BACKUP_FILE_NAME									string			= 	"backup.txt"

	// Elevator controller

	SHOULD_DISPLAY_WORKERS								bool 			= 	false

	DISTRIBUTOR_ALIVE_NOTIFICATION_DELAY 				time.Duration 	=	time.Millisecond * 40
	DISTRIBUTOR_CONNECTION_CHECK_DELAY 					time.Duration 	=	time.Millisecond * 80
				
	TIMEOUT_TIME_ORDER_TAKEN							time.Duration	= 	time.Millisecond * 100

	TIMEOUT_TIME_NOT_FUNCTIONAL							time.Duration	= 	time.Second * 5

	// Elevator

	SHOULD_DISPLAY_ELEVATOR 							bool 			= 	false

	NUMBER_OF_FLOORS		 							int 			= 	4

	REGISTER_ELEVATOR_EVENT_SLEEP						time.Duration 	= 	time.Millisecond * 10
		
	ELEVATOR_DOOR_OPEN_DURATION 						time.Duration 	= 	time.Second * 3
);