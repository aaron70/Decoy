package decoy

import (
	"encoding/base64"
	"encoding/json"
)

func (d *Decoy) JsonUnmarshalBytes(data []byte) (map[string]any, error) {
	var res map[string]any
	err := json.Unmarshal(data, &res)
	return res, err
}

func (d *Decoy) JsonUnmarshalString(data string) (map[string]any, error) {
	return d.JsonUnmarshalBytes([]byte(data))
}

func (d *Decoy) Base64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
