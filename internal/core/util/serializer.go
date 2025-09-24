package util

import "encoding/json"

func Serialize(data any) ([]byte, error) {
	return json.Marshal(data)
}

func Deserialize(data []byte, output any) error {
	return json.Unmarshal(data, output)
}
