package repository

import (
	"encoding/base64"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func encodeCursor(k map[string]types.AttributeValue) string {
	if len(k) == 0 {
		return ""
	}
	var m map[string]string
	attributevalue.UnmarshalMap(k, &m)
	b, _ := json.Marshal(m)
	return base64.URLEncoding.EncodeToString(b)
}

func decodeCursor(cursor string) (map[string]types.AttributeValue, error) {
	if cursor == "" {
		return nil, nil
	}
	b, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	var m map[string]string
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}
	return attributevalue.MarshalMap(m)
}
