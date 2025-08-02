package constanta

// ReportStatus mewakili status pemrosesan laporan.
type ReportStatus string

const (
	StatusPending    ReportStatus = "PENDING"
	StatusInProgress ReportStatus = "IN_PROGRESS"
	StatusCompleted  ReportStatus = "COMPLETED"
	StatusFailed     ReportStatus = "FAILED"
)
