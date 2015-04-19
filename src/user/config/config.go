package config;

import(
	"time"
);

const(
	MASTER_ALIVE_NOTIFICATION_DELAY 	time.Duration 	=	time.Millisecond * 40

	NUMBER_OF_FLOORS		 			int 			= 	4

	SHOULD_DISPLAY_ELEVATOR 			bool 			= 	false
	SHOULD_DISPLAY_WORKERS				bool 			= 	false

	PORT_SERVER_DEFAULT 				int 			= 	9125
	PORT_SERVER_WITH_TIMEOUT			int 			= 	9126
);