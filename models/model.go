package models

import (
	"als-coding-test-2/constanta"
	"time"
)

// --- Struktur Data ---
// ReportRequest mewakili permintaan untuk membuat laporan.
type ReportRequest struct {
	ID         string            `json:"id"`
	ReportType string            `json:"report_type"` // e.g., "sales", "inventory"
	Parameters map[string]string `json:"parameters"`
	CreatedAt  time.Time         `json:"created_at"`
}

type ReportResult struct {
	RequestID   string                 `json:"request_id"`
	Status      constanta.ReportStatus `json:"status"`
	GeneratedAt time.Time              `json:"generated_at"`
	ReportData  string                 `json:"report_data,omitempty"` // Contoh data laporan (bisa complex struct)
	Error       string                 `json:"error,omitempty"`
}
