package commands

import (
	"database/sql"
	"fmt"
	"os"
	"proxychan/internal/server"
	"time"
)

type ConnView struct {
	ID          uint64
	Username    string
	SourceIP    string
	Destination string
	StartedAt   time.Time
}

func runListConnections(db *sql.DB) {
	groups, err := server.ListActiveConnectionsByIP()
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}

	if len(groups) == 0 {
		fmt.Println("no active connections")
		return
	}

	now := time.Now()

	for _, g := range groups {
		fmt.Printf(
			"SOURCE %s (%d connections)\n",
			g.SourceIP,
			g.Count,
		)

		for _, c := range g.Conns {
			age := now.Sub(c.StartedAt).Truncate(time.Second)

			user := c.Username
			if user == "" {
				user = "-"
			}

			fmt.Printf(
				"  ID=%d USER=%s DST=%s AGE=%s\n",
				c.ID,
				user,
				c.Destination,
				age,
			)
		}

		fmt.Println()
	}
}
