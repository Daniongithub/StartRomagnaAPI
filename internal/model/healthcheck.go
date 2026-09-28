package model

type ReplicationStatus struct {
	MasterHost string `json:"master_host"`
	IORunning  bool   `json:"io_running"`
	SQLRunning bool   `json:"sql_running"`
	LagSeconds *int   `json:"lag_seconds"` // nil se NULL (thread IO non connesso)
	CaughtUp   bool   `json:"caught_up"`
}

type DBStatus struct {
	Role         string             `json:"role"` // master | replica | resyncing | broken
	ReadOnly     bool               `json:"read_only"`
	GtidPosition string             `json:"gtid_position"`
	Replication  *ReplicationStatus `json:"replication"`
}
