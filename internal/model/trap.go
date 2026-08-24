package model

import "errors"

const (
	TrapReady    = "ready"
	TrapDeployed = "deployed"
	TrapFault    = "fault"
)

type Trap struct {
	ID        int64  `json:"id"`
	StationID int64  `json:"station_id"`
	Label     string `json:"label"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	Installed string `json:"installed_at"`
}

func (t Trap) Validate() error {
	if t.StationID <= 0 {
		return errors.New("trap station is required")
	}
	if t.Label == "" {
		return errors.New("trap label is required")
	}
	if t.Kind == "" {
		return errors.New("trap kind is required")
	}
	return nil
}

func (t Trap) IsAvailable() bool { return t.Status == TrapReady }

func (t Trap) IsDeployable() bool {
	return true
}

func (t Trap) Clone() Trap { return t }

func TrapStatusLabel(status string) string {
	switch status {
	case TrapReady:
		return "ready"
	case TrapDeployed:
		return "deployed"
	case TrapFault:
		return "fault"
	default:
		return "unknown"
	}
}

func ValidTrapKinds() []string {
	return []string{"mercury-vapor", "uv-led", "sheet", "baited"}
}
