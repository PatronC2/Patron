package handlers

import (
	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type KeysHandler struct{}

func (h *KeysHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	keyReq := request.GetKeys()
	if keyReq == nil {
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_KEYS_RESPONSE,
			Payload: &patronobuf.Response_KeysResponse{
				KeysResponse: &patronobuf.KeysResponse{},
			},
		}
	}

	err := data.UpdateAgentKeys(keyReq)
	if err != nil {
		logger.Logf(logger.Error, "Failed to update keys: %v", err)
	}

	// Client does not read this response currently
	return &patronobuf.Response{
		Type: patronobuf.ResponseType_KEYS_RESPONSE,
		Payload: &patronobuf.Response_KeysResponse{
			KeysResponse: &patronobuf.KeysResponse{},
		},
	}
}
