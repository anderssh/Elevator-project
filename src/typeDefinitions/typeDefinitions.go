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

type ButtonType int

const(
	BUTTON_CALL_UP     	ButtonType = iota
	BUTTON_CALL_DOWN   	ButtonType = iota
	BUTTON_CALL_INSIDE 	ButtonType = iota

	BUTTON_STOP         ButtonType = iota
	BUTTON_OBSTRUCTION  ButtonType = iota
);

//-----------------------------------------------//

type ButtonFloor struct {
	
	Type       				ButtonType
	Floor 					int
	
	IoRegisterPressed 		int
	PressedReadingPrevious	bool
	PressedReadingCurrent	bool
	
	IoRegisterLight			int
}

func (button *ButtonFloor) UpdateState() {

	button.PressedReadingPrevious = button.PressedReadingCurrent;
	button.PressedReadingCurrent = io.IsBitSet(button.IoRegisterPressed);
}

func (button *ButtonFloor) IsPressed() bool {
	return button.PressedReadingCurrent && !button.PressedReadingPrevious;
}

func (button *ButtonFloor) TurnOffLight() {
	io.ClearBit(button.IoRegisterLight);
}

func (button *ButtonFloor) TurnOnLight() {
	io.SetBit(button.IoRegisterLight);
}

//-----------------------------------------------//

type ButtonSimple struct {
	Type       				ButtonType
	
	IoRegisterPressed 		int
	PressedReadingPrevious	bool
	PressedReadingCurrent	bool
}

func (button *ButtonSimple) UpdateState() {

	button.PressedReadingPrevious = button.PressedReadingCurrent;
	button.PressedReadingCurrent = io.IsBitSet(button.IoRegisterPressed);
}

func (button *ButtonSimple) IsPressed() bool {
	return (button.PressedReadingCurrent && !button.PressedReadingPrevious);
}

func (button *ButtonSimple) IsReleased() bool {
	return (!button.PressedReadingCurrent && button.PressedReadingPrevious);
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