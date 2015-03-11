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
		fmt.Print("\x1b[31;1m");
		fmt.Print(value);
		fmt.Println("\x1b[0m");
	}
}

func Warning(value interface{}) {
	
	if logLevel >= LOG_WARNING {
		fmt.Print("\x1b[33;1m");
		fmt.Print(value);
		fmt.Println("\x1b[0m");
	}
}

func Data(value interface{}) {

	if logLevel >= LOG_ALL {
		fmt.Println(value);
	}
}