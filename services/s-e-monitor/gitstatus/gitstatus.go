// Package gitstatus liest den Git-Status eines Repos über die git-CLI.
package gitstatus

// Zuletzt geändert: 2026-08-14

import (
	"fmt"
	"os/exec"
	"strings"
)

// Status ist der Git-Zustand eines Repos (Schema: event.git.status).
type Status struct {
	RepoPath    string
	Branch      string
	Staged      int
	Uncommitted int
	Ahead       int
	Behind      int
}

// Params liefert die params für event.git.status.
func (s Status) Params() map[string]any {
	return map[string]any{
		"repo_path":   s.RepoPath,
		"branch":      s.Branch,
		"staged":      s.Staged,
		"uncommitted": s.Uncommitted,
		"ahead":       s.Ahead,
		"behind":      s.Behind,
	}
}

// Git liest den Status über ausführbare git-Aufrufe (injezierbar für Tests).
type Git struct {
	Dir  string
	Exec func(args ...string) (string, error)
}

// New erstellt einen Git-Statusleser für das angegebene Verzeichnis.
func New(dir string) *Git {
	return &Git{Dir: dir, Exec: execGit(dir)}
}

func execGit(dir string) func(args ...string) (string, error) {
	return func(args ...string) (string, error) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		out, err := cmd.Output()
		return strings.TrimSpace(string(out)), err
	}
}

// Read ermittelt den aktuellen Status. Einzelne Fehler degradieren: Branch
// fällt auf den kurzen Hash zurück, Zähler bleiben 0 bei fehlendem Upstream.
func (g *Git) Read() (Status, error) {
	s := Status{RepoPath: g.Dir}

	if out, err := g.Exec("symbolic-ref", "--short", "-q", "HEAD"); err == nil && out != "" {
		s.Branch = out
	} else if hash, err := g.Exec("rev-parse", "--short", "HEAD"); err == nil {
		s.Branch = hash
	} else {
		return s, fmt.Errorf("kein git-Repo in %s", g.Dir)
	}

	if out, err := g.Exec("status", "--porcelain=v1"); err == nil {
		s.Staged, s.Uncommitted = countPorcelain(out)
	}

	if out, err := g.Exec("rev-list", "--left-right", "--count", "HEAD...@{upstream}"); err == nil {
		s.Ahead, s.Behind = parseCount(out)
	}

	return s, nil
}

func countPorcelain(out string) (staged, uncommitted int) {
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		if x == '?' && y == '?' { // untracked
			uncommitted++
			continue
		}
		if x != ' ' {
			staged++
		}
		if y != ' ' {
			uncommitted++
		}
	}
	return staged, uncommitted
}

func parseCount(out string) (ahead, behind int) {
	fields := strings.Fields(out)
	if len(fields) == 2 {
		fmt.Sscanf(fields[0], "%d", &ahead)
		fmt.Sscanf(fields[1], "%d", &behind)
	}
	return ahead, behind
}
