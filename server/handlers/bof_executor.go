package handlers

import (
	"fmt"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/data"
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
					Uuid:      "",
					Success:   false,
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
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{
					Uuid:         agentUUID,
					Success:      false,
					FunctionName: err.Error(),
				},
			},
		}
	}

	if len(pendingTasks) == 0 {
		logger.Logf(logger.Info, "No pending BOF tasks for agent %s", agentUUID)
		
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
			Payload: &patronobuf.Response_BofExecuteResponse{
				BofExecuteResponse: &patronobuf.BoFExecuteResponse{
					Uuid:    agentUUID,
					Success: true,
					Output:  "No pending BOF tasks",
				},
			},
		}
	}

	logger.Logf(logger.Info, "Found %d pending BOF task(s) for agent %s", len(pendingTasks), agentUUID)

	// Process each pending task and send BOF execute requests
	for _, task := range pendingTasks {
		// Get BOF payload
		bofMetadata, bofPayload, err := data.GetBOFByUuid(task.BofUuid, "windows") // TODO: Use task.Platform if available
		if err != nil {
			logger.Logf(logger.Error, "Failed to get BOF %s for agent %s: %v", task.BofUuid, agentUUID, err)
			continue
		}

		// Send BOF execute request to agent with payload
		bofReq := &patronobuf.Request{
			Type: patronobuf.RequestType_BOF_EXECUTE,
			Payload: &patronobuf.Request_BofExecute{
				BofExecute: &patronobuf.BoFExecuteRequest{
					Uuid:         agentUUID,
					Platform:     "windows", // TODO: Use task.Platform
					FunctionName: bofMetadata.FunctionName,
					Payload:      bofPayload,
				},
			},
		}

		if err := common.WriteDelimited(stream, bofReq); err != nil {
			logger.Logf(logger.Error, "Failed to send BOF request to agent %s: %v", agentUUID, err)
			break
		}

		logger.Logf(logger.Info, "Sent BOF execute request for function %s to agent %s", task.BofUuid, agentUUID)
		
		data.MarkTaskCompleted(task.Id, agentUUID, "completed", "BOF sent successfully")
	}

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_BOF_EXECUTE_RESPONSE,
		Payload: &patronobuf.Response_BofExecuteResponse{
			BofExecuteResponse: &patronobuf.BoFExecuteResponse{
				Uuid:    agentUUID,
				Success: true,
				Output:  fmt.Sprintf("Sent %d BOF task(s)", len(pendingTasks)),
			},
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
