package handlers

import (
	"net"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
	"github.com/PatronC2/Patron/lib/logger"
)

// Returns the next command to run to an agent. Updates the DB after the command gets run

type CommandHandler struct{}

func getNextCommand(c types.CommandRequest) (types.CommandResponse, error) {
	fetch, err := data.FetchNextCommand(c.AgentID)
	if err != nil {
		logger.Logf(logger.Warning, "Couldn't fetch agent: %v\n", err)
		return types.CommandResponse{}, false
	}

	return CommandResponse, nil
}

func (h *CommandHandler) Handle(request types.Request, conn net.Conn) types.Response {
    commandReq, ok := request.Payload.(types.CommandRequest)
    if !ok {
        return types.Response{
            Type:    types.CommandResponseType,
            Payload: types.CommandResponse{},
        }
    }

    configResponse, success := getNextCommand(commandReq)
    if !success {
        return types.Response{
            Type:    types.CommandResponseType,
            Payload: types.CommandResponse{},
        }
    }

    return types.Response{
        Type:    types.CommandResponseType,
        Payload: configResponse,
    }
}

func (h *CommandStatusHandler) Handle(request types.Request, conn net.Conn) types.Response {
    c, ok := request.Payload.(types.CommandStatusRequest)
	data.UpdateAgentCommand(c.CommandID, c.CommandOutput, c.AgentID)
	// This type doesn't actually matter since the client won't read it
	return types.Response{
		Type:		types.CommandRequest,
		Payload:	types.CommandRequest{},
	}
}
