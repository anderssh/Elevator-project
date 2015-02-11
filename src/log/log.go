package log;

import (
	"fmt"
);

//-----------------------------------------------//

const (
	LOG_LEVEL_ERROR 	int 	= 	1
	LOG_LEVEL_DEBUG 	int 	= 	2
);

var logLevel = LOG_LEVEL_DEBUG;

func SetLogLevel(newLogLevel int) {
	logLevel = newLogLevel;
}

//-----------------------------------------------//

func Error(value interface{}) {
	if logLevel >= LOG_LEVEL_ERROR {
		fmt.Println(value);
	}
}

func Debug(value interface{}) {
	if logLevel >= LOG_LEVEL_DEBUG {
		fmt.Println(value);
	}
}