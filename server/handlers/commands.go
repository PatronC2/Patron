package handlers

import (
	"net"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
)

// Returns the next command to run to an agent. Updates the DB after the command gets run

type CommandHandler struct{}
type CommandStatusHandler struct{}

func (h *CommandHandler) Handle(request types.Request, conn net.Conn) types.Response {
    commandReq, ok := request.Payload.(types.CommandRequest)
    if !ok {
        return types.Response{
            Type:    types.CommandResponseType,
            Payload: types.CommandResponse{},
        }
    }
    CommandResponse := data.FetchNextCommand(commandReq.AgentID)
    return types.Response{
        Type:    types.CommandResponseType,
        Payload: CommandResponse,
    }
}

func (h *CommandStatusHandler) Handle(request types.Request, conn net.Conn) types.Response {
    c, ok := request.Payload.(types.CommandStatusRequest)
    if !ok {
        return types.Response{
            Type:    types.CommandStatusResponseType,
            Payload: types.CommandStatusResponse{},
        }
    }
	data.UpdateAgentCommand(c.CommandID, c.CommandResult, c.CommandOutput, c.AgentID)
	// This type doesn't actually matter since the client won't read it
	return types.Response{
		Type:    types.CommandStatusResponseType,
		Payload: types.CommandStatusResponse{},
	}
}
