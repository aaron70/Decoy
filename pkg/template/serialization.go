package template

import (
	"encoding/base64"
)

func (t *Template) Base64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}

func (t *Template) Base64String(value string) string {
	return base64.StdEncoding.EncodeToString([]byte(value))
}

func (t *Template) Base64Decode(value string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(value)
}
