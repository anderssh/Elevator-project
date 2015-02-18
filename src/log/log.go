package log;

import (
	"fmt"
);

//-----------------------------------------------//

const (
	LOG_NONE 		int 	= 		iota
	LOG_ERROR 		int 	= 		iota
	LOG_WARNING		int 	= 		iota
	LOG_ALL			int 	= 		iota
);

var logLevel = LOG_ALL;

func SetLogLevel(newLogLevel int) {
	logLevel = newLogLevel;
}

//-----------------------------------------------//

func Error(value interface{}) {
	if logLevel >= LOG_ERROR {
		fmt.Println(value);
	}
}

func Warning(value interface{}) {
	if logLevel >= LOG_WARNING {
		fmt.Println(value);
	}
}

func Data(value interface{}) {
	if logLevel >= LOG_ALL {
		fmt.Println(value);
	}
}