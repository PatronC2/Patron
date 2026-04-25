package api

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/PatronC2/Patron/data"
	"github.com/PatronC2/Patron/lib/logger"
	"github.com/gin-gonic/gin"
)

func ListFilesForUUIDHandler(c *gin.Context) {
	uuid := c.Param("agt")
	logger.Logf(logger.Info, "Listing files for %v", uuid)

	files, err := data.ListFilesForUUID(uuid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get files"})
		return
	}

	fileList := []gin.H{}
	for _, file := range files {
		fileList = append(fileList, gin.H{
			"FileID": file.FileID,
			"Path":   file.Path,
			"Status": file.Status,
		})
	}

	c.JSON(http.StatusOK, gin.H{"data": fileList})
}

func DownloadFileHandler(c *gin.Context) {
	fileID := c.Param("fileid")
	logger.Logf(logger.Info, "Downloading file %v", fileID)

	content, filename, err := data.DownloadFile(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download file"})
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Data(http.StatusOK, "application/octet-stream", content)
}

func UploadFileHandler(c *gin.Context) {
	transfertype := c.PostForm("transfertype")

	// If the transfer type is "Upload", skip the file content
	if transfertype == "Upload" {
		path := c.PostForm("path")
		uuid := c.PostForm("uuid")

		err := data.UploadFile(path, uuid, transfertype, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "Uploaded successfully"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to retrieve file"})
		return
	}

	path := c.PostForm("path")
	uuid := c.PostForm("uuid")

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file content"})
		return
	}

	err = data.UploadFile(path, uuid, transfertype, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Uploaded successfully"})
}

// Handles the /api/files/list GET endpoint
func ListFilesHandler(c *gin.Context) {
	logger.Logf(logger.Debug, "Listing all files")

	tagFilters := c.QueryArray("tag")
	logic := c.DefaultQuery("logic", "or")
	limitStr := c.DefaultQuery("limit", "20")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	files, total, nextOffset, err := data.ListFiles(tagFilters, logic, limit, offset)

	if err != nil {
		logger.Logf(logger.Error, "Error listing all files: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list all files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       files,
		"totalCount": total,
		"nextOffset": nextOffset,
	})
}
