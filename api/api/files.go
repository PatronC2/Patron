package api

import (
	"fmt"
	"io"
	"net/http"

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
			"Type":   file.Type,
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

	err = data.UploadFile(path, uuid, content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "Uploaded successfully"})
}


