//go:build linux
// +build linux

package linux_utils

import (
	"io"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
)

// HandleBOFExecuteRequest requests and processes BOF execution tasks from server
func HandleBOFExecuteRequest(beacon io.ReadWriteCloser, agentID string) error {
	logger.Logf(logger.Info, "Fetching BOFs to execute")

	// Agent sends request to server asking for BOF tasks
	req := &patronobuf.Request{
		Type: patronobuf.RequestType_BOF_EXECUTE,
		Payload: &patronobuf.Request_BofExecute{
			BofExecute: &patronobuf.BoFExecuteRequest{
				Uuid: agentID,
			},
		},
	}

	if err := common.WriteDelimited(beacon, req); err != nil {
		logger.Logf(logger.Error, "Error sending BOF execute request: %v", err)
		return err
	}

	// Read response from server
	resp := &patronobuf.Response{}
	if err := common.ReadDelimited(beacon, resp); err != nil {
		logger.Logf(logger.Error, "Error reading BOF execute response: %v", err)
		return err
	}

	bofResp := resp.GetBofExecuteResponse()
	if bofResp == nil {
		logger.Logf(logger.Warning, "Invalid BOF execute response")
		return nil
	}

	logger.Logf(logger.Info, "BOF Execute Response - UUID: %s, Success: %v, Output: %s",
		bofResp.GetUuid(), bofResp.GetSuccess(), bofResp.GetOutput())

	if !bofResp.GetSuccess() {
		logger.Logf(logger.Error, "BOF execution failed")
		return nil
	}

	// TODO: Execute the BOF payload if provided
	// For now, just acknowledge receipt
	return nil
}
