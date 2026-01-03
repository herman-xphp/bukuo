package handler

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload endpoints
type UploadHandler struct {
	uploadDir string
	baseURL   string
	maxSize   int64 // Maximum file size in bytes
}

// allowedMimeTypes maps MIME types to their allowed extensions
var allowedMimeTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// Magic bytes for image file validation
var magicBytes = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8, 0xFF},
	"image/png":  {0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
	"image/gif":  {0x47, 0x49, 0x46, 0x38}, // GIF8
	"image/webp": {0x52, 0x49, 0x46, 0x46}, // RIFF (WebP starts with RIFF)
}

// NewUploadHandler creates a new UploadHandler
func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	// Ensure upload directory exists with restrictive permissions
	os.MkdirAll(uploadDir, 0750)
	return &UploadHandler{
		uploadDir: uploadDir,
		baseURL:   baseURL,
		maxSize:   2 * 1024 * 1024, // 2MB max (reduced from 5MB for security)
	}
}

// Upload handles POST /uploads
// @Summary Upload a file
// @Description Upload an image file (JPEG, PNG, GIF, WebP only, max 2MB)
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file to upload"
// @Success 200 {object} map[string]string "url and filename"
// @Failure 400 {object} map[string]string "error message"
// @Failure 500 {object} map[string]string "error message"
// @Security BearerAuth
// @Router /api/upload [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	// 1. Parse multipart form with size limit
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.maxSize+512) // +512 for form overhead

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "File too large. Maximum size is 2MB"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}
	defer file.Close()

	// 2. Validate file size (double check)
	if header.Size > h.maxSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size must be less than 2MB"})
		return
	}

	// 3. Read first 512 bytes for content detection
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read file"})
		return
	}
	buffer = buffer[:n]

	// 4. Detect actual MIME type from content (not from Content-Type header which can be spoofed)
	detectedType := http.DetectContentType(buffer)

	// 5. Validate against allowed types
	ext, allowed := allowedMimeTypes[detectedType]
	if !allowed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only JPEG, PNG, GIF, and WebP images are allowed"})
		return
	}

	// 6. Verify magic bytes for extra security
	if !verifyMagicBytes(buffer, detectedType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image file"})
		return
	}

	// 7. Generate secure random filename (ignore original filename completely)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate filename"})
		return
	}
	filename := fmt.Sprintf("%s_%d%s", hex.EncodeToString(randomBytes), time.Now().UnixNano(), ext)

	// 8. Create random subdirectory to prevent enumeration
	subDir := hex.EncodeToString(randomBytes[:2]) // First 2 bytes = 4 hex chars
	dirPath := filepath.Join(h.uploadDir, subDir)
	if err := os.MkdirAll(dirPath, 0750); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}

	filePath := filepath.Join(dirPath, filename)

	// 9. Ensure the final path is within upload directory (prevent path traversal)
	absUploadDir, _ := filepath.Abs(h.uploadDir)
	absFilePath, _ := filepath.Abs(filePath)
	if !strings.HasPrefix(absFilePath, absUploadDir) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	// 10. Create file with restrictive permissions
	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	defer out.Close()

	// 11. Write the buffer we already read
	if _, err := out.Write(buffer); err != nil {
		os.Remove(filePath) // Cleanup on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 12. Copy the rest of the file
	if _, err := io.Copy(out, file); err != nil {
		os.Remove(filePath) // Cleanup on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// 13. Return URL
	fileURL := fmt.Sprintf("%s/uploads/%s/%s", h.baseURL, subDir, filename)
	c.JSON(http.StatusOK, gin.H{
		"url":      fileURL,
		"filename": filename,
	})
}

// verifyMagicBytes checks if the file content matches expected magic bytes for the MIME type
func verifyMagicBytes(content []byte, mimeType string) bool {
	expected, exists := magicBytes[mimeType]
	if !exists {
		return false
	}

	if len(content) < len(expected) {
		return false
	}

	// Special handling for WebP - it starts with RIFF, then 4 bytes, then WEBP
	if mimeType == "image/webp" {
		if len(content) < 12 {
			return false
		}
		// Check RIFF header
		if !bytes.HasPrefix(content, expected) {
			return false
		}
		// Check WEBP signature at offset 8
		if !bytes.Equal(content[8:12], []byte("WEBP")) {
			return false
		}
		return true
	}

	return bytes.HasPrefix(content, expected)
}

// SecureStaticHandler returns a handler that serves static files with security headers
func SecureStaticHandler(root string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Security headers to prevent XSS and other attacks
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'none'; img-src 'self'")
		c.Header("Cache-Control", "public, max-age=31536000, immutable") // Cache for 1 year (files are immutable)

		// Serve the file
		fullPath := filepath.Join(root, c.Param("filepath"))

		// Prevent path traversal
		absRoot, _ := filepath.Abs(root)
		absPath, _ := filepath.Abs(fullPath)
		if !strings.HasPrefix(absPath, absRoot) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		// Check if file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.File(fullPath)
	}
}
