package system

import (
	"database/sql"
	"fmt"
	"net"
)

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
