package share

// GlobalShareRequest represents a request to add a file/folder to the global share.
type GlobalShareRequest struct {
	ID        uint   `json:"id" storm:"id,increment"`
	Path      string `json:"path" storm:"index"`
	UserID    uint   `json:"userID" storm:"index"`
	Username  string `json:"username,omitempty"`
	Status    string `json:"status" storm:"index"` // "pending", "approved", "rejected"
	CreatedAt int64  `json:"createdAt"`
	Message   string `json:"message,omitempty"` // Optional message from user
}

// GlobalShare represents a file/folder in the global share.
type GlobalShare struct {
	ID           uint   `json:"id" storm:"id,increment"`
	Path         string `json:"path" storm:"index"` // Path in global share folder
	OriginalPath string `json:"originalPath"`       // Original path before promotion
	UserID       uint   `json:"userID"`             // User who shared it
	Username     string `json:"username,omitempty"`
	AddedAt      int64  `json:"addedAt"`
	RequestID    uint   `json:"requestID,omitempty"` // Reference to original request
}
