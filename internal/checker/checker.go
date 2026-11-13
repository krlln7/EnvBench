package checker

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
)

type Report struct {
	Verdict string           `json:"verdict"`
	Levels  map[string]Level `json:"levels"`
}

type Level struct {
	Status string   `json:"status"`
	Diffs  []string `json:"diffs,omitempty"`
}

func EqEnv(ctx context.Context, envDSN, referenceDSN string) (Report, error) {
	rep := Report{Verdict: "equivalent", Levels: map[string]Level{}}

	// L1
	l1, err := checkSchema(ctx, envDSN, referenceDSN)
	if err != nil {
		return rep, err
	}
	rep.Levels["L1"] = l1
	if l1.Status != "ok" {
		rep.Verdict = "not_equivalent"
	}

	// needs L2

	// L3
	l3, err := checkService(ctx, envDSN, referenceDSN)
	if err != nil {
		return rep, err
	}
	rep.Levels["L3"] = l3
	if l3.Status != "ok" {
		rep.Verdict = "not_equivalent"
	}

	return rep, nil
}

func checkSchema(ctx context.Context, a, b string) (Level, error) {
	da, err := dumpSchema(ctx, a)
	if err != nil {
		return Level{}, err
	}
	db, err := dumpSchema(ctx, b)
	if err != nil {
		return Level{}, err
	}
	na := normalizeSchema(da)
	nb := normalizeSchema(db)
	if na == nb {
		return Level{Status: "ok"}, nil
	}
	return Level{Status: "diff", Diffs: diffLines(na, nb)}, nil
}

func dumpSchema(ctx context.Context, dsn string) (string, error) {
	cmd := exec.CommandContext(ctx, "pg_dump",
		"--schema-only", "--no-owner", "--no-privileges", "--no-comments", dsn)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("pg_dump: %w", err)
	}
	return string(out), nil
}

// sort lines, remove spacebars, comments
func normalizeSchema(s string) string {
	var lines []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "--") {
			continue
		}
		lines = append(lines, l)
	}
	sort.Strings(lines)
	h := sha256.Sum256([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(h[:])
}

func diffLines(a, b string) []string {
	if a == b {
		return nil
	}
	return []string{fmt.Sprintf("schema hash differs: env=%s ref=%s", a[:12], b[:12])}
}

func checkService(ctx context.Context, a, b string) (Level, error) {
	va, err := serviceSnapshot(ctx, a)
	if err != nil {
		return Level{}, err
	}
	vb, err := serviceSnapshot(ctx, b)
	if err != nil {
		return Level{}, err
	}
	ja, _ := json.Marshal(va)
	jb, _ := json.Marshal(vb)
	if string(ja) == string(jb) {
		return Level{Status: "ok"}, nil
	}
	return Level{Status: "diff", Diffs: []string{
		fmt.Sprintf("env=%s", ja),
		fmt.Sprintf("ref=%s", jb),
	}}, nil
}

func serviceSnapshot(ctx context.Context, dsn string) (map[string]any, error) {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	defer conn.Close(ctx)

	out := map[string]any{}

	// sequences
	rows, err := conn.Query(ctx, `SELECT sequencename, increment_by, start_value FROM pg_sequences ORDER BY sequencename`)
	if err != nil {
		return nil, err
	}
	var seqs []map[string]any
	for rows.Next() {
		var name string
		var inc, start int64
		if err := rows.Scan(&name, &inc, &start); err != nil {
			return nil, err
		}
		seqs = append(seqs, map[string]any{"name": name, "inc": inc, "start": start})
	}
	rows.Close()
	out["sequences"] = seqs

	// extensions
	rows, err = conn.Query(ctx, `SELECT extname, extversion FROM pg_extension ORDER BY extname`)
	if err != nil {
		return nil, err
	}
	var exts []map[string]any
	for rows.Next() {
		var name, ver string
		if err := rows.Scan(&name, &ver); err != nil {
			return nil, err
		}
		exts = append(exts, map[string]any{"name": name, "version": ver})
	}
	rows.Close()
	out["extensions"] = exts

	// parameters
	rows, err = conn.Query(ctx, `SELECT name, setting FROM pg_settings
		WHERE name IN ('default_transaction_isolation','timezone','lc_collate','server_version')
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	params := map[string]string{}
	for rows.Next() {
		var n, v string
		if err := rows.Scan(&n, &v); err != nil {
			return nil, err
		}
		params[n] = v
	}
	rows.Close()
	out["settings"] = params

	return out, nil
}
