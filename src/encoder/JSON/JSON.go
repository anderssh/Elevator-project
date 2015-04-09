package JSON;

import "encoding/json"

//-----------------------------------------------//

func Encode(v interface{}) ([]byte, error){
	return json.Marshal(v);
}

func Decode(data []byte, v interface{}) (error){ 		// Use address of interface, e.g. &order
	return json.Unmarshal(data, v);
}