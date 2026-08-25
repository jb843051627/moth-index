package model

func IsFinalBatchStatus(status string) bool { return status == BatchClosed || status == BatchArchived }

func IsActiveSpecimenStatus(status string) bool {
	return status == SpecimenCaptured || status == SpecimenPending
}

func IsRecoverableTaskStatus(status string) bool { return status == TaskQueued || status == TaskFailed }

func SeverityLabel(value string) string {
	switch value {
	case SeverityBlocker:
		return "阻断"
	case SeverityWarn:
		return "警告"
	case SeverityInfo:
		return "提示"
	default:
		return "未知"
	}
}

func StatusGroup(value string) string {
	switch value {
	case StationActive, TrapReady, BatchOpen, SpecimenCaptured, ReviewPending, TaskQueued:
		return "active"
	case StationRetired, TrapFault, BatchArchived, SpecimenRejected, TaskFailed:
		return "closed"
	default:
		return "transition"
	}
}
