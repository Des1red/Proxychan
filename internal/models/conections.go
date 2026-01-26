package models

import "time"

type ActiveConn struct {
	ID          uint64    `json:"id"`
	Username    string    `json:"username"`
	SourceIP    string    `json:"source_ip"`
	Destination string    `json:"destination"`
	StartedAt   time.Time `json:"started_at"`
}

type ConnGroup struct {
	SourceIP string       `json:"source_ip"`
	Count    int          `json:"count"`
	Conns    []ActiveConn `json:"conns"`
}
