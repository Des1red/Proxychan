package system

import (
	"database/sql"
	"fmt"
	"net"
)

// Runtime function only
func LoadWhitelist(db *sql.DB) ([]net.IPNet, error) {
	rows, err := db.Query(`SELECT cidr FROM whitelist WHERE enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []net.IPNet
	for rows.Next() {
		var cidr string
		if err := rows.Scan(&cidr); err != nil {
			return nil, err
		}
		_, netw, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid cidr %q: %w", cidr, err)
		}
		out = append(out, *netw)
	}
	return out, rows.Err()
}

func GetWhitelistVersion(db *sql.DB) (int64, error) {
	var v int64
	err := db.QueryRow(
		`SELECT version FROM whitelist_meta WHERE id = 1`,
	).Scan(&v)
	return v, err
}
func BumpWhitelistVersion(db *sql.DB) error {
	_, err := db.Exec(
		`UPDATE whitelist_meta SET version = version + 1 WHERE id = 1`,
	)
	return err
}

func normalizeCIDR(input string) (string, error) {
	if ip := net.ParseIP(input); ip != nil {
		// single IP → /32 or /128
		if ip.To4() != nil {
			return ip.String() + "/32", nil
		}
		return ip.String() + "/128", nil
	}

	_, _, err := net.ParseCIDR(input)
	if err != nil {
		return "", fmt.Errorf("invalid IP/CIDR: %s", input)
	}
	return input, nil
}

// add if not exist , if exist grant access
func AllowIP(db *sql.DB, input string) error {
	cidr, err := normalizeCIDR(input)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO whitelist (cidr, enabled)
		VALUES (?, 1)
		ON CONFLICT(cidr) DO UPDATE SET enabled = 1
	`, cidr)

	return err
}

// revoke access
func BlockIP(db *sql.DB, input string) error {
	cidr, err := normalizeCIDR(input)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		UPDATE whitelist SET enabled = 0 WHERE cidr = ?
	`, cidr)

	return err
}

// hard remove
func DeleteIP(db *sql.DB, input string) error {
	cidr, err := normalizeCIDR(input)
	if err != nil {
		return err
	}

	res, err := db.Exec(`DELETE FROM whitelist WHERE cidr = ?`, cidr)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("ip not found: %s", cidr)
	}

	return nil
}

type WhitelistEntry struct {
	CIDR    string
	Enabled bool
}

func ListWhitelist(db *sql.DB) ([]WhitelistEntry, error) {
	rows, err := db.Query(`SELECT cidr, enabled FROM whitelist ORDER BY cidr`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []WhitelistEntry
	for rows.Next() {
		var e WhitelistEntry
		var enabled int
		if err := rows.Scan(&e.CIDR, &enabled); err != nil {
			return nil, err
		}
		e.Enabled = enabled == 1
		out = append(out, e)
	}
	return out, rows.Err()
}

type WhitelistStatus struct {
	Version  int64
	Enabled  int
	Disabled int
	Total    int
}

func GetWhitelistStatus(db *sql.DB) (*WhitelistStatus, error) {
	var s WhitelistStatus

	if err := db.QueryRow(
		`SELECT version FROM whitelist_meta WHERE id = 1`,
	).Scan(&s.Version); err != nil {
		return nil, err
	}

	row := db.QueryRow(`
		SELECT
			COUNT(*) AS total,
			SUM(CASE WHEN enabled = 1 THEN 1 ELSE 0 END) AS enabled,
			SUM(CASE WHEN enabled = 0 THEN 1 ELSE 0 END) AS disabled
		FROM whitelist
	`)

	if err := row.Scan(&s.Total, &s.Enabled, &s.Disabled); err != nil {
		return nil, err
	}

	return &s, nil
}

func ClearWhitelist(db *sql.DB) error {
	_, err := db.Exec(`
		UPDATE whitelist
		SET enabled = 0
		WHERE cidr NOT IN ('127.0.0.1/32', '::1/128')
	`)
	return err
}
