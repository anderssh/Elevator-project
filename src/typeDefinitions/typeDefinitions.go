package typeDefinitions;

import(
	"../io"
);

//-----------------------------------------------//

type Direction int

const(
	DIRECTION_UP   		Direction 	= iota
	DIRECTION_DOWN 		Direction 	= iota
);

//-----------------------------------------------//

type OrderType int 

const(
	ORDER_UP 		OrderType = iota
	ORDER_DOWN 		OrderType = iota
	ORDER_INSIDE 	OrderType = iota
);

type Order struct {
	Type 	OrderType
	Floor	int
}