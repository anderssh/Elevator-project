package io;

/*
#cgo LDFLAGS: -lcomedi -lm
#include "io.h"
*/
import "C";

//-----------------------------------------------//

const (
	// In ports
	PORT4 					int 	= 	3
	OBSTRUCTION 			int		=   (0x300+23)
	STOP            		int		=   (0x300+22)
	BUTTON_CALL_INSIDE1 	int		=   (0x300+21)
	BUTTON_CALL_INSIDE2 	int 	=	(0x300+20)
	BUTTON_CALL_INSIDE3 	int 	=	(0x300+19)
	BUTTON_CALL_INSIDE4 	int 	=	(0x300+18)
	BUTTON_UP1      		int 	=	(0x300+17)
	BUTTON_UP2      		int 	=	(0x300+16)

	PORT1               	int 	= 	2
	BUTTON_DOWN2        	int 	= 	(0x200+0)
	BUTTON_UP3          	int 	= 	(0x200+1)
	BUTTON_DOWN3        	int 	= 	(0x200+2)
	BUTTON_DOWN4        	int 	= 	(0x200+3)
	SENSOR_FLOOR1      	 	int 	= 	(0x200+4)
	SENSOR_FLOOR2       	int 	= 	(0x200+5)
	SENSOR_FLOOR3       	int 	= 	(0x200+6)
	SENSOR_FLOOR4       	int 	= 	(0x200+7)

	// Out ports
	PORT3	            	int   	= 	3
	MOTORDIR	        	int   	= 	(0x300+15)
	LIGHT_STOP	        	int 	= 	(0x300+14)
	LIGHT_CALL_INSIDE1	    int 	= 	(0x300+13)
	LIGHT_CALL_INSIDE2	    int 	= 	(0x300+12)
	LIGHT_CALL_INSIDE3	    int 	= 	(0x300+11)
	LIGHT_CALL_INSIDE4	    int 	= 	(0x300+10)
	LIGHT_UP1	        	int 	= 	(0x300+9)
	LIGHT_UP2	        	int 	= 	(0x300+8)

	PORT2               	int 	= 	3
	LIGHT_DOWN2         	int 	= 	(0x300+7)
	LIGHT_UP3           	int 	= 	(0x300+6)
	LIGHT_DOWN3         	int 	= 	(0x300+5)
	LIGHT_DOWN4         	int 	= 	(0x300+4)
	LIGHT_DOOR_OPEN     	int 	= 	(0x300+3)
	LIGHT_FLOOR_INDICATOR2  int 	= 	(0x300+1)
	LIGHT_FLOOR_INDICATOR1  int 	= 	(0x300+0)

	PORT0               	int 	= 	1
	MOTOR               	int 	= 	(0x100+0)

	// Non existing ports
	BUTTON_DOWN1        	int 	= 	-1
	BUTTON_UP4          	int 	= 	-1
	LIGHT_DOWN1         	int 	= 	-1
	LIGHT_UP4           	int 	= 	-1
);

//-----------------------------------------------//

type ErrorIO struct {
    errorMessage string
}

func (e *ErrorIO) Error() string {
    return "IO error: " + e.errorMessage
}

//-----------------------------------------------//

func Initialize() *ErrorIO {
	
	err := C.io_init();

	if err == 0 {
		return &ErrorIO{"Failed to initialize hardware."};
	}
	
	return nil
}

//-----------------------------------------------//

func SetBit(busChannel int){
	C.io_set_bit(C.int(busChannel));	
}

func ClearBit(busChannel int){
	C.io_clear_bit(C.int(busChannel));	
}

func IsBitSet(busChannel int) bool { 

	if int(C.io_read_bit(C.int(busChannel))) == 1 {
		return true;
	}

	return false;
}

//-----------------------------------------------//

func WriteAnalog(busChannel int, value int){
	C.io_write_analog(C.int(busChannel), C.int(value));
}

func ReadAnalog(busChannel int) int {
	return int(C.io_read_analog(C.int(busChannel)));
}