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

func Error(values ... interface{}) {
	
	if logLevel >= LOG_ERROR {
		
		fmt.Print("\x1b[31;1m");	// Set color to red
		fmt.Print("Error: ");

		for _, value := range values {
			fmt.Print(value);
			fmt.Print(" ");
		}

		fmt.Println("\x1b[0m");  	// Reset color
	}
}

func Warning(values ... interface{}) {
	
	if logLevel >= LOG_WARNING {
	
		fmt.Print("\x1b[33;1m");	// Set color to yellow
		fmt.Print("Warning: ");

		for _, value := range values {
			fmt.Print(value);
			fmt.Print(" ");
		}
		
		fmt.Println("\x1b[0m");  	// Reset color
	}
}

func Data(values ... interface{}) {

	if logLevel >= LOG_ALL {

		for _, value := range values {
			fmt.Print(value);
			fmt.Print(" ");
		}

		fmt.Println("");
	}
}