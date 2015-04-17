package config;

import(
	"time"
);

const(
	MASTER_ALIVE_NOTIFICATION_DELAY 	time.Duration 	=	time.Millisecond * 100
	MASTER_ALIVE_NOTIFICATION_TIMEOUT 	time.Duration 	=	time.Millisecond * 400

	SLAVE_ALIVE_NOTIFICATION_DELAY 		time.Duration 	=	time.Millisecond * 40
	SLAVE_ALIVE_NOTIFICATION_TIMEOUT 	time.Duration 	=	time.Millisecond * 200

	TIMEOUT_TIME_COST_RESPONSE 			time.Duration 	=	time.Millisecond * 200	

	SHOULD_DISPLAY_ELEVATOR 			bool 			= 	false
	SHOULD_DISPLAY_WORKERS				bool 			= 	true
);