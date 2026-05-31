package grpcx

import (
	"encoding/json"

	"google.golang.org/grpc/encoding"
)

const JSONCodecName = "json"

type JSONCodec struct{}

func init() {
	encoding.RegisterCodec(JSONCodec{})
}

func (JSONCodec) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONCodec) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSONCodec) Name() string {
	return JSONCodecName
}
