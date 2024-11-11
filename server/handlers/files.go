package handlers

import (
	"net"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
)

type FileRequestHandler struct{}

func (h *FileRequestHandler) Handle(request types.Request, conn net.Conn) types.Response {
    commandReq, ok := request.Payload.(types.FileRequest)
    if !ok {
        return types.Response{
            Type:    types.FileResponseType,
            Payload: types.FileResponse{},
        }
    }
    FileResponse := data.FetchNextFileTransfer(commandReq.AgentID)
    return types.Response{
        Type:    types.FileResponseType,
        Payload: FileResponse,
    }
}
