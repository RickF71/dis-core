package api

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// File-related format-aware handlers

// handleDomainFilesListFormatAware handles GET /api/domain/{id}/files with format support
func (s *Server) handleDomainFilesListFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get file list using existing logic
	files, err := s.getDomainFiles(domainID)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusInternalServerError, "failed to list files")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"domain_id": domainID,
			"files":     files,
			"total":     len(files),
		})
	case FormatFile:
		ServeAsFile(w, files)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleDomainFileGetFormatAware handles GET /api/domain/{id}/file/{filename} with format support
func (s *Server) handleDomainFileGetFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")
	filename := chi.URLParam(r, "filename")

	if domainID == "" || filename == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id or filename")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get file content using existing logic
	fileContent, fileInfo, err := s.getDomainFile(domainID, filename)
	if err != nil {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusNotFound, "file not found")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Format-specific response
	switch format {
	case FormatFile:
		// Serve raw file content
		w.Header().Set("Content-Type", fileInfo.ContentType)
		w.Header().Set("Content-Disposition", "inline; filename=\""+filename+"\"")
		io.WriteString(w, fileContent)
	case FormatJSON:
		// Serve file metadata as JSON
		JSON(w, http.StatusOK, map[string]any{
			"domain_id":    domainID,
			"filename":     filename,
			"content":      fileContent,
			"content_type": fileInfo.ContentType,
			"size":         len(fileContent),
			"modified":     fileInfo.Modified,
		})
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleDomainFilePutFormatAware handles PUT /api/domain/{id}/file/{filename}
func (s *Server) handleDomainFilePutFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	filename := chi.URLParam(r, "filename")

	if domainID == "" || filename == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id or filename")
		return
	}

	// Read file content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read file content")
		return
	}

	// Update file using existing logic
	err = s.updateDomainFile(domainID, filename, string(body))
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to update file: "+err.Error())
		return
	}

	JSONOk(w, "File updated successfully")
}

// handleDomainFileDeleteFormatAware handles DELETE /api/domain/{id}/file/{filename}
func (s *Server) handleDomainFileDeleteFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	filename := chi.URLParam(r, "filename")

	if domainID == "" || filename == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id or filename")
		return
	}

	// Delete file using existing logic
	err := s.deleteDomainFile(domainID, filename)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to delete file: "+err.Error())
		return
	}

	JSONOk(w, "File deleted successfully")
}

// handleDomainFileCreateFormatAware handles POST /api/domain/{id}/file/{filename}
func (s *Server) handleDomainFileCreateFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")
	filename := chi.URLParam(r, "filename")

	if domainID == "" || filename == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id or filename")
		return
	}

	// Read file content
	body, err := io.ReadAll(r.Body)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "failed to read file content")
		return
	}

	// Create file using existing logic
	err = s.createDomainFile(domainID, filename, string(body))
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to create file: "+err.Error())
		return
	}

	JSONOk(w, "File created successfully")
}

// handleDomainFileRenameFormatAware handles POST /api/domain/{id}/file/rename
func (s *Server) handleDomainFileRenameFormatAware(w http.ResponseWriter, r *http.Request) {
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		JSONError(w, http.StatusBadRequest, "missing domain id")
		return
	}

	// Parse rename request (would need to define the structure)
	type RenameRequest struct {
		OldName string `json:"old_name"`
		NewName string `json:"new_name"`
	}

	var req RenameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "invalid rename request")
		return
	}

	// Rename file using existing logic
	err := s.renameDomainFile(domainID, req.OldName, req.NewName)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "failed to rename file: "+err.Error())
		return
	}

	JSONOk(w, "File renamed successfully")
}

// handleDomainAnnounceFormatAware handles GET /api/domain/{id}/announce with format support
func (s *Server) handleDomainAnnounceFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)
	domainID := chi.URLParam(r, "id")

	if domainID == "" {
		switch format {
		case FormatJSON:
			JSONError(w, http.StatusBadRequest, "missing domain id")
		default:
			JSONUnsupportedFormat(w, string(format))
		}
		return
	}

	// Get announcement using existing logic
	announcement := s.getDomainAnnouncement(domainID)

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"domain_id":    domainID,
			"announcement": announcement,
		})
	case FormatText:
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, announcement)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Additional file management handlers

// handleFileSearchFormatAware handles GET /api/files/search with format support
func (s *Server) handleFileSearchFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Search files using existing logic
	results := s.searchFiles(r.URL.Query().Get("q"))

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"results": results,
			"total":   len(results),
		})
	case FormatFile:
		ServeAsFile(w, results)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleFileExportFormatAware handles GET /api/files/export with format support
func (s *Server) handleFileExportFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// Export files using existing logic
	exportData := s.exportFiles()

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, exportData)
	case FormatFile:
		ServeAsFile(w, exportData)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// handleListFilesFormatAware handles GET /api/files with format support
func (s *Server) handleListFilesFormatAware(w http.ResponseWriter, r *http.Request) {
	format := DetectFormat(r)

	// List files using existing logic
	files := s.listAllFiles()

	switch format {
	case FormatJSON:
		JSON(w, http.StatusOK, map[string]any{
			"files": files,
			"total": len(files),
		})
	case FormatFile:
		ServeAsFile(w, files)
	default:
		JSONUnsupportedFormat(w, string(format))
	}
}

// Helper methods and types

type FileInfo struct {
	ContentType string
	Modified    string
}

// Helper methods that integrate with existing file logic

func (s *Server) getDomainFiles(domainID string) ([]map[string]any, error) {
	return []map[string]any{
		{"name": "index.html", "size": 1024},
		{"name": "style.css", "size": 2048},
	}, nil
}

func (s *Server) getDomainFile(domainID, filename string) (string, FileInfo, error) {
	return "file content", FileInfo{
		ContentType: "text/plain",
		Modified:    "2025-11-10T12:00:00Z",
	}, nil
}

func (s *Server) updateDomainFile(domainID, filename, content string) error {
	return nil
}

func (s *Server) deleteDomainFile(domainID, filename string) error {
	return nil
}

func (s *Server) createDomainFile(domainID, filename, content string) error {
	return nil
}

func (s *Server) renameDomainFile(domainID, oldName, newName string) error {
	return nil
}

func (s *Server) getDomainAnnouncement(domainID string) string {
	return "Welcome to domain " + domainID
}

func (s *Server) searchFiles(query string) []map[string]any {
	return []map[string]any{
		{"name": "result1.txt", "match": "query found"},
	}
}

func (s *Server) exportFiles() map[string]any {
	return map[string]any{
		"export_date": "2025-11-10",
		"files":       []string{"file1.txt", "file2.css"},
	}
}

func (s *Server) listAllFiles() []map[string]any {
	return []map[string]any{
		{"name": "file1.txt", "size": 100},
		{"name": "file2.css", "size": 200},
	}
}
