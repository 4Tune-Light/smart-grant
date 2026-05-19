package cursor

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Cursor struct {
	LastID        string    `json:"last_id"`
	LastCreatedAt time.Time `json:"last_created_at"`
}

func Encode(c Cursor) string {
	b, _ := json.Marshal(c)
	return base64.URLEncoding.EncodeToString(b)
}

func Decode(s string) (Cursor, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, err
	}
	var c Cursor
	err = json.Unmarshal(b, &c)
	return c, err
}
