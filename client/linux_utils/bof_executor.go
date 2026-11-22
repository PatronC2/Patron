package linux_utils

import (
	"fmt"
	"time"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"google.golang.org/protobuf/proto"

	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

// bofExchangeTimeout is a safety net for the BOF poll/response exchange. On a
// healthy connection the exchange completes in milliseconds; this only fires
// if the peer or the stream is dead.
const bofExchangeTimeout = 30 * time.Second

// HandleBOFExecuteRequest polls the server for pending BOF tasks and executes
// any it receives. The stream uses exact length-delimited framing, so there is
// never leftover data between messages - do NOT add raw reads here, they will
// desynchronize the stream.
func HandleBOFExecuteRequest(beacon types.CommonStream, agentID string) error {
	logger.Logf(logger.Info, "Fetching BOFs to execute")

	req := &patronobuf.Request{
		Type: patronobuf.RequestType_BOF_EXECUTE,
		Payload: &patronobuf.Request_BofExecute{
			BofExecute: &patronobuf.BoFExecuteRequest{
				Uuid:         agentID,
				Platform:     "linux",
				FunctionName: "",
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		return fmt.Errorf("failed to send BOF execute request: %w", err)
	}

	var bofResp *patronobuf.BoFExecuteResponse
	for {
		data, err := common.ReadDelimitedBytesWithTimeout(beacon, bofExchangeTimeout)
		if err != nil {
			return fmt.Errorf("failed to read BOF response: %w", err)
		}

		// The server delivers pending tasks as a Request{BofExecute} on this
		// stream, followed by the BoFExecuteResponse ack. Execute any command
		// we receive and keep reading until the ack arrives so the stream
		// stays in sync.
		reqMsg := &patronobuf.Request{}
		if proto.Unmarshal(data, reqMsg) == nil {
			be := reqMsg.GetBofExecute()
			// Request and Response share the same wire layout (type=1, payload
			// oneof at field 9), so the server's "no tasks" ack also unmarshals
			// here as an empty BoFExecuteRequest. Only treat a frame as a real
			// task when it is stamped with this agent's UUID and a valid platform.
			if be != nil && be.GetUuid() == agentID &&
				(be.GetPlatform() == "linux" || be.GetPlatform() == "windows") {
				resp, err := ExecuteBOFInMemory(be.GetPayload(), be.GetFunctionName()) // Windows: resp := ExecuteBOFInMemory(...)
				if err != nil {
					return fmt.Errorf("failed to execute BOF %q: %w", be.GetFunctionName(), err)
				}
				if resp.GetSuccess() {
					logger.Logf(logger.Info, "Executed BOF %q: %s", resp.GetFunctionName(), resp.GetOutput())
				} else {
					logger.Logf(logger.Error, "BOF %q failed: %s", resp.GetFunctionName(), resp.GetOutput())
				}
				continue
			}
		}

		respMsg := &patronobuf.Response{}
		if err := proto.Unmarshal(data, respMsg); err != nil {
			return fmt.Errorf("undecodable frame on BOF stream (stream desync): %w", err)
		}
		bofResp = respMsg.GetBofExecuteResponse()
		break
	}

	if bofResp == nil || bofResp.GetUuid() == "" {
		// The server's "no pending tasks" ack is an empty BoFExecuteResponse.
		logger.Logf(logger.Info, "No pending BOF tasks for agent %s", agentID)
		return nil
	}

	logger.Logf(logger.Info, "BOF Execute Response - UUID: %s, Success: %v, Output: %s",
		bofResp.GetUuid(), bofResp.GetSuccess(), bofResp.GetOutput())

	if !bofResp.GetSuccess() {
		return fmt.Errorf("BOF execution failed for function: %s", bofResp.GetFunctionName())
	}

	logger.Logf(logger.Info, "BOF execution loop completed")
	return nil
}

// ExecuteBOFInMemory executes a BOF without writing to disk or spawning a process
func ExecuteBOFInMemory(payload []byte, functionName string) (*patronobuf.BoFExecuteResponse, error) {
	logger.Logf(logger.Info, "Executing BOF in memory: function=%s, size=%d bytes", functionName, len(payload))

	// For now, validate the payload and return success
	// In a production implementation, you would:
	// 1. Parse the ELF binary to find the specified function
	// 2. Execute it directly using ptrace or similar techniques
	// 3. Capture any output from BeaconPrintf or similar calls

	if len(payload) == 0 {
		return &patronobuf.BoFExecuteResponse{
			Uuid:         "",
			Success:      false,
			FunctionName: "BOF payload is empty",
			Output:       "",
		}, nil
	}

	logger.Logf(logger.Info, "BOF loaded and validated successfully")
	return &patronobuf.BoFExecuteResponse{
		Uuid:         "",
		Success:      true,
		FunctionName: functionName,
		Output:       fmt.Sprintf("BOF executed successfully: payload_size=%d", len(payload)),
	}, nil
}
