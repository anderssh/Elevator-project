package config;

import(
	"time"
);

const(
	DISTRIBUTOR_ALIVE_NOTIFICATION_DELAY 				time.Duration 	=	time.Millisecond * 40
				
	REGISTER_ELEVATOR_EVENT_SLEEP						time.Duration 	= 	time.Millisecond * 10
				
	NUMBER_OF_FLOORS		 							int 			= 	4
	ELEVATOR_DOOR_OPEN_DURATION 						time.Duration 	= 	time.Second * 3
				
	TIMEOUT_TIME_ORDER_TAKEN							time.Duration	= 	time.Millisecond * 100
				
	SHOULD_DISPLAY_ELEVATOR 							bool 			= 	false
	SHOULD_DISPLAY_WORKERS								bool 			= 	false
				
	PORT_SERVER_DEFAULT 								int 			= 	9125
	PORT_SERVER_WITH_TIMEOUT							int 			= 	9126

	SHOULD_USE_PROCESS_PAIRS 							bool 			= 	true

	BACKUP_PROCESS_ALIVE_MESSAGE_DEADLINE				time.Duration	=   time.Millisecond * 200
	BACKUP_PROCESS_ALIVE_NOTIFICATION_SLEEP  			time.Duration	=   time.Millisecond * 15
);