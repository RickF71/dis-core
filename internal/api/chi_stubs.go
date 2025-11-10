package api

import (
	"net/http"
)

// Stub handlers for chi router format-aware endpoints
// These are placeholder implementations that return "not yet implemented" errors

// Domain handlers
func (s *Server) handleCreateDomainChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "create domain format-aware handler not yet implemented")
}

func (s *Server) handleUpdateDomainChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "update domain format-aware handler not yet implemented")
}

func (s *Server) handleUpdateDomainCSSChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "update domain CSS format-aware handler not yet implemented")
}

// Domain file handlers
func (s *Server) handleDomainFilesListChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain files list format-aware handler not yet implemented")
}

func (s *Server) handleDomainFileGetChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain file get format-aware handler not yet implemented")
}

func (s *Server) handleDomainFilePutChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain file put format-aware handler not yet implemented")
}

func (s *Server) handleDomainFileDeleteChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain file delete format-aware handler not yet implemented")
}

func (s *Server) handleDomainFileCreateChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain file create format-aware handler not yet implemented")
}

func (s *Server) handleDomainFileRenameChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain file rename format-aware handler not yet implemented")
}

// Domain policy handlers
func (s *Server) handleGetDomainPolicyChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "get domain policy format-aware handler not yet implemented")
}

func (s *Server) handleSetDomainPolicyChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "set domain policy format-aware handler not yet implemented")
}

// Domain announcement handlers
func (s *Server) handleDomainAnnounceChiFormatAware(w http.ResponseWriter, r *http.Request) {
	JSONUnsupportedFormat(w, "domain announce format-aware handler not yet implemented")
}

// Core API handlers - these exist in handlers_status_chi.go
// Commenting out the duplicates:
// func (s *Server) handlePingFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleHealthFormatAware(w http.ResponseWriter, r *http.Request)

// Policy handlers - some methods already exist in handlers_policy_chi.go
// Commenting out the duplicates:
// func (s *Server) handleListPoliciesFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleGetPolicyFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handlePutPolicyFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleDeletePolicyFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handlePolicyReloadFormatAware(w http.ResponseWriter, r *http.Request)

// Status handlers - these exist in handlers_status_chi.go
// Commenting out the duplicates:
// func (s *Server) handleStatusFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleAuthorityStatusFormatAware(w http.ResponseWriter, r *http.Request)

// File handlers - some methods already exist in handlers_file_chi.go
// Commenting out the duplicates:
// func (s *Server) handleFileSearchFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleFileExportFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleListFilesFormatAware(w http.ResponseWriter, r *http.Request)

// Receipt handlers - some methods already exist in handlers_receipt_chi.go
// Commenting out the duplicates:
// func (s *Server) handleListReceiptsFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleGetReceiptFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleReceiptSearchFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleVerifyReceiptFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleCreateReceiptFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleDeleteReceiptFormatAware(w http.ResponseWriter, r *http.Request)

// Version handlers - these exist in handlers_receipt_chi.go
// Commenting out the duplicates:
// func (s *Server) handleListVersionsFormatAware(w http.ResponseWriter, r *http.Request)
// func (s *Server) handleVersionExportFormatAware(w http.ResponseWriter, r *http.Request)
