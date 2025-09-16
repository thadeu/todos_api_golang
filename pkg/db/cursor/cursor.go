package cursor

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"strings"
)

type CursorData struct {
	Datetime string `json:"datetime"`
	ID       int    `json:"id,omitempty"`
}

type CursorResponse struct {
	Size       int             `json:"size"`
	Data       json.RawMessage `json:"data"`
	Pagination struct {
		HasNext    bool   `json:"has_next"`
		NextCursor string `json:"next_cursor"`
	} `json:"pagination"`
}

func hmacSignature(encoded string) string {
	mac := hmac.New(sha256.New, []byte(os.Getenv("CURSOR_SECRET_KEY")))
	mac.Write([]byte(encoded))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func verifySignature(encoded string, signature string) bool {
	expectedSignature := hmacSignature(encoded)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

func EncodeCursor(date string, id int) string {
	data := CursorData{Datetime: date, ID: id}
	jsonData, _ := json.Marshal(data)
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	signature := hmacSignature(encoded)

	return encoded + "." + signature
}

func DecodeCursor(token string) (string, int, error) {
	parts := strings.Split(token, ".")

	if len(parts) != 2 {
		return "", 0, errors.New("invalid cursor format")
	}

	if !verifySignature(parts[0], parts[1]) {
		return "", 0, errors.New("invalid cursor signature")
	}

	decoded, err := base64.StdEncoding.DecodeString(parts[0])

	if err != nil {
		return "", 0, err
	}

	var cursor CursorData
	json.Unmarshal(decoded, &cursor)

	return cursor.Datetime, cursor.ID, nil
}
