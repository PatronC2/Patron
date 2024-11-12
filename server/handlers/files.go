package handlers

import (
	"net"

    "github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/types"
)

type FileRequestHandler struct{}
type FileToServerHandler struct{}

func (h *FileRequestHandler) Handle(request types.Request, conn net.Conn) types.Response {
    fileReq, ok := request.Payload.(types.FileRequest)
    if !ok {
        return types.Response{
            Type:    types.FileResponseType,
            Payload: types.FileResponse{},
        }
    }
    FileResponse := data.FetchNextFileTransfer(fileReq.AgentID)
    return types.Response{
        Type:    types.FileResponseType,
        Payload: FileResponse,
    }
}

func (h *FileToServerHandler) Handle(request types.Request, conn net.Conn) types.Response {
    fileTransfer, ok := request.Payload.(types.FileToServer)
    if !ok {
        return types.Response{
            Type:    types.FileTransferStatusResponseType,
            Payload: types.FileTransferStatusResponse{},
        }
    }
    err := data.UpdateFileTransfer(fileTransfer)
    if err != nil {
        logger.Logf(logger.Error, "Error updating file in db: %v", err)
    }    
    return types.Response{
        Type:    types.FileTransferStatusResponseType,
        Payload: types.FileTransferStatusResponse{},
    }
}
