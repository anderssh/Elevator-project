package config;

import(
	"time"
);

const(
	MASTER_ALIVE_NOTIFICATION_DELAY 	time.Duration 	=	time.Millisecond * 100
	MASTER_ALIVE_NOTIFICATION_TIMEOUT 	time.Duration 	=	time.Millisecond * 400

	SLAVE_ALIVE_NOTIFICATION_DELAY 		time.Duration 	=	time.Millisecond * 40
	SLAVE_ALIVE_NOTIFICATION_TIMEOUT 	time.Duration 	=	time.Millisecond * 200
);