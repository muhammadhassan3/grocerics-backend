package domain

import "time"

// QCRawResponse is one QuickCommerce HTTP call, captured raw. Append-only, and
// never read by the application -- it exists to hand the ML person real QC data.
type QCRawResponse struct {
	BaseModel
	Endpoint     string    `gorm:"not null" json:"endpoint"`
	Params       string    `gorm:"type:jsonb;not null" json:"params"`
	StatusCode   int       `gorm:"not null" json:"status_code"`
	Response     *string   `gorm:"type:jsonb" json:"response,omitempty"`
	ResponseText *string   `json:"response_text,omitempty"`
	Error        *string   `json:"error,omitempty"`
	DurationMs   int       `gorm:"not null" json:"duration_ms"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func (QCRawResponse) TableName() string { return "qc_raw_responses" }
