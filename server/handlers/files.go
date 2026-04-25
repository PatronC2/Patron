package handlers

import (
	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/PatronC2/Patron/types"
)

type FileRequestHandler struct{}
type FileToServerHandler struct{}

func (h *FileRequestHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	req := request.GetFile()
	if req == nil {
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_FILE_RESPONSE,
			Payload: &patronobuf.Response_FileResponse{
				FileResponse: &patronobuf.FileResponse{},
			},
		}
	}

	resp := data.FetchNextFileTransfer(req.GetUuid())
	if resp == nil {
		resp = &patronobuf.FileResponse{}
	}

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_FILE_RESPONSE,
		Payload: &patronobuf.Response_FileResponse{
			FileResponse: resp,
		},
	}
}

func (h *FileToServerHandler) Handle(request *patronobuf.Request, stream types.CommonStream) *patronobuf.Response {
	file := request.GetFileToServer()
	if file == nil {
		return &patronobuf.Response{
			Type: patronobuf.ResponseType_FILE_TRANSFER_STATUS,
			Payload: &patronobuf.Response_FileTransferStatusResponse{
				FileTransferStatusResponse: &patronobuf.FileTransferStatusResponse{},
			},
		}
	}

	err := data.UpdateFileTransfer(file)
	if err != nil {
		logger.Logf(logger.Error, "Error updating file transfer: %v", err)
	}

	return &patronobuf.Response{
		Type: patronobuf.ResponseType_FILE_TRANSFER_STATUS,
		Payload: &patronobuf.Response_FileTransferStatusResponse{
			FileTransferStatusResponse: &patronobuf.FileTransferStatusResponse{},
		},
	}
}
