package template

import (
	"encoding/base64"
)

func (t *Template) Base64(bytes []byte) string {
	return base64.StdEncoding.EncodeToString(bytes)
}
