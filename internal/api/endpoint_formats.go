package api

// EndpointFormats defines the allowed formats for specific endpoints
var EndpointFormats = map[string][]Format{
	"/api/domain/{id}/css":  {FormatFile, FormatText, FormatJSON},
	"/api/receipts/list":    {FormatJSON, FormatCBOR},
	"/api/bootstrap/status": {FormatJSON, FormatText},
	"/api/ping":             {FormatJSON, FormatText, FormatFile},
	"/api/health":           {FormatJSON, FormatText},
	"/api/status":           {FormatJSON, FormatText},
}

// GetAllowedFormats returns the allowed formats for a given endpoint pattern
func GetAllowedFormats(pattern string) ([]Format, bool) {
	formats, exists := EndpointFormats[pattern]
	return formats, exists
}
