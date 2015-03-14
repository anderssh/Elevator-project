package typeDefinitions;

//-----------------------------------------------//

type Direction int

const(
	DIRECTION_UP   		Direction 	= iota
	DIRECTION_DOWN 		Direction 	= iota
);

//-----------------------------------------------//

type ButtonType int

const(
	BUTTON_CALL_UP     	ButtonType = iota
	BUTTON_CALL_DOWN   	ButtonType = iota
	BUTTON_CALL_INSIDE 	ButtonType = iota

	BUTTON_STOP         ButtonType = iota
	BUTTON_OBSTRUCTION  ButtonType = iota
);

type ButtonFloor struct {
	Type       			ButtonType
	Floor 				int
	Pressed				bool
	Light				bool
	BusChannelPressed 	int
	BusChannelLight	int
}

type ButtonSimple struct {
	Type       			ButtonType
	Pressed    			bool
	Light				bool
	BusChannelPressed 	int
}

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