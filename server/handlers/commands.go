package handlers

import (
	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type CommandHandler struct{}
type CommandStatusHandler struct{}

func (h *CommandHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	commandReq := request.GetCommand()
	if commandReq == nil {
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_COMMAND_RESPONSE,
			Payload: &patronobuf.Response_CommandResponse{
				CommandResponse: &patronobuf.CommandResponse{},
			},
		}
	}

	command := data.FetchNextCommand(commandReq.GetUuid())

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_COMMAND_RESPONSE,
		Payload: &patronobuf.Response_CommandResponse{
			CommandResponse: command,
		},
	}
}

func (h *CommandStatusHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	status := request.GetCommandStatus()
	if status == nil {
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_COMMAND_STATUS_RESPONSE,
			Payload: &patronobuf.Response_CommandStatusResponse{
				CommandStatusResponse: &patronobuf.CommandStatusResponse{},
			},
		}
	}

	err := data.UpdateAgentCommand(
		status.GetCommandid(),
		status.GetResult(),
		status.GetOutput(),
		status.GetUuid(),
	)
	if err != nil {
		logger.Logf(logger.Error, "Failed to update command: %v", err)
	}

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_COMMAND_STATUS_RESPONSE,
		Payload: &patronobuf.Response_CommandStatusResponse{
			CommandStatusResponse: &patronobuf.CommandStatusResponse{},
		},
	}
}
