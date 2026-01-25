package system

import (
	"database/sql"
	"fmt"
	"net"
	"strings"
)

type DenyType string

const (
	DenyIP          DenyType = "ip"
	DenyCIDR        DenyType = "cidr"
	DenyDomainExact DenyType = "domain_exact"
	DenyDomainSuf   DenyType = "domain_suffix"
)

type DenyRule struct {
	Pattern string
	Type    DenyType
	Enabled bool
}

// ---------- versioning (mirror whitelist) ----------

func GetDenylistVersion(db *sql.DB) (int64, error) {
	var v int64
	err := db.QueryRow(`SELECT version FROM denylist_meta WHERE id = 1`).Scan(&v)
	return v, err
}

func BumpDenylistVersion(db *sql.DB) error {
	_, err := db.Exec(`UPDATE denylist_meta SET version = version + 1 WHERE id = 1`)
	return err
}

// ---------- normalization ----------

// DNS is case-insensitive. Also tolerate trailing dot.
func normalizeDomain(d string) (string, error) {
	d = strings.TrimSpace(d)
	if d == "" {
		return "", fmt.Errorf("empty domain")
	}
	d = strings.ToLower(d)
	d = strings.TrimSuffix(d, ".")
	if d == "" {
		return "", fmt.Errorf("invalid domain")
	}
	return d, nil
}

func classifyAndNormalizePattern(input string) (pattern string, typ DenyType, err error) {
	in := strings.TrimSpace(input)
	if in == "" {
		return "", "", fmt.Errorf("empty pattern")
	}

	// IP?
	if ip := net.ParseIP(in); ip != nil {
		// store canonical string
		return ip.String(), DenyIP, nil
	}

	// CIDR?
	if strings.Contains(in, "/") {
		if _, n, e := net.ParseCIDR(in); e == nil {
			return n.String(), DenyCIDR, nil
		}
		// fallthrough; could be a domain with slash (rare) but we don't support that anyway
	}

	// Domain suffix: must start with dot
	if strings.HasPrefix(in, ".") {
		d, e := normalizeDomain(in[1:])
		if e != nil {
			return "", "", e
		}
		return "." + d, DenyDomainSuf, nil
	}

	// Exact domain
	d, e := normalizeDomain(in)
	if e != nil {
		return "", "", e
	}
	return d, DenyDomainExact, nil
}

// ---------- CRUD ----------

// DenyDestination enables (or inserts) a deny rule.
func DenyDestination(db *sql.DB, input string) error {
	pattern, typ, err := classifyAndNormalizePattern(input)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
		INSERT INTO denylist (pattern, type, enabled)
		VALUES (?, ?, 1)
		ON CONFLICT(pattern) DO UPDATE SET enabled = 1, type = excluded.type
	`, pattern, string(typ))
	if err != nil {
		return err
	}

	return BumpDenylistVersion(db)
}

// AllowDestination disables a deny rule (soft remove).
func AllowDestination(db *sql.DB, input string) error {
	pattern, _, err := classifyAndNormalizePattern(input)
	if err != nil {
		return err
	}

	_, err = db.Exec(`UPDATE denylist SET enabled = 0 WHERE pattern = ?`, pattern)
	if err != nil {
		return err
	}

	return BumpDenylistVersion(db)
}

// DeleteDestination hard-deletes a rule.
func DeleteDestination(db *sql.DB, input string) error {
	pattern, _, err := classifyAndNormalizePattern(input)
	if err != nil {
		return err
	}

	res, err := db.Exec(`DELETE FROM denylist WHERE pattern = ?`, pattern)
	if err != nil {
		return err
	}

	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("deny rule not found: %s", pattern)
	}

	return BumpDenylistVersion(db)
}

func ListDenylist(db *sql.DB) ([]DenyRule, error) {
	rows, err := db.Query(`SELECT pattern, type, enabled FROM denylist ORDER BY type, pattern`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []DenyRule
	for rows.Next() {
		var r DenyRule
		var enabled int
		var typ string
		if err := rows.Scan(&r.Pattern, &typ, &enabled); err != nil {
			return nil, err
		}
		r.Type = DenyType(typ)
		r.Enabled = enabled == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// Runtime: load enabled rules and pre-parse.
type DenylistRuntime struct {
	IPNets       []net.IPNet
	DomainExact  map[string]struct{}
	DomainSuffix []string // stored like ".example.com"
}

func LoadDenylist(db *sql.DB) (*DenylistRuntime, error) {
	rows, err := db.Query(`SELECT pattern, type FROM denylist WHERE enabled = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rt := &DenylistRuntime{
		DomainExact: make(map[string]struct{}),
	}

	for rows.Next() {
		var p string
		var typ string
		if err := rows.Scan(&p, &typ); err != nil {
			return nil, err
		}

		switch DenyType(typ) {
		case DenyIP:
			ip := net.ParseIP(p)
			if ip == nil {
				return nil, fmt.Errorf("invalid deny ip pattern in db: %q", p)
			}
			if ip.To4() != nil {
				_, n, _ := net.ParseCIDR(ip.String() + "/32")
				rt.IPNets = append(rt.IPNets, *n)
			} else {
				_, n, _ := net.ParseCIDR(ip.String() + "/128")
				rt.IPNets = append(rt.IPNets, *n)
			}

		case DenyCIDR:
			_, n, err := net.ParseCIDR(p)
			if err != nil {
				return nil, fmt.Errorf("invalid deny cidr in db: %q: %w", p, err)
			}
			rt.IPNets = append(rt.IPNets, *n)

		case DenyDomainExact:
			d, err := normalizeDomain(p)
			if err != nil {
				return nil, fmt.Errorf("invalid deny domain exact in db: %q: %w", p, err)
			}
			rt.DomainExact[d] = struct{}{}

		case DenyDomainSuf:
			if !strings.HasPrefix(p, ".") {
				return nil, fmt.Errorf("invalid deny domain suffix in db: %q", p)
			}
			d, err := normalizeDomain(p[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid deny domain suffix in db: %q: %w", p, err)
			}
			rt.DomainSuffix = append(rt.DomainSuffix, "."+d)

		default:
			return nil, fmt.Errorf("unknown deny type in db: %q", typ)
		}
	}

	return rt, rows.Err()
}

func ClearDenylist(db *sql.DB) error {
	_, err := db.Exec(`
		UPDATE denylist
		SET enabled = 0
	`)
	return err
}
