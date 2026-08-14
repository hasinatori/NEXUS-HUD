package gitstatus

// Zuletzt geändert: 2026-08-14

import (
	"os"
	"os/exec"
	"testing"
)

func initRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	return dir
}

func commit(t *testing.T, dir, file, content string) {
	t.Helper()
	os.WriteFile(dir+"/"+file, []byte(content), 0o644)
	run := func(args ...string) {
		cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, out)
		}
	}
	run("add", file)
	run("commit", "-q", "-m", "add "+file)
}

func TestGitStatusClean(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "a.txt", "a")
	g := New(dir)
	s, err := g.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if s.Branch != "master" && s.Branch != "main" {
		t.Fatalf("branch=%q, want master/main", s.Branch)
	}
	if s.Staged != 0 || s.Uncommitted != 0 {
		t.Fatalf("staged=%d uncommitted=%d, want 0/0", s.Staged, s.Uncommitted)
	}
	if s.Ahead != 0 || s.Behind != 0 {
		t.Fatalf("ahead=%d behind=%d, want 0/0 (kein Upstream)", s.Ahead, s.Behind)
	}
}

func TestGitStatusCounts(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "a.txt", "a")
	commit(t, dir, "b.txt", "b")
	os.WriteFile(dir+"/mod.txt", []byte("neu"), 0o644)
	cmd := exec.Command("git", "-C", dir, "add", "mod.txt")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("add: %v (%s)", err, out)
	}
	os.WriteFile(dir+"/mod.txt", []byte("neu2"), 0o644) // staged + worktree-Änderung -> AM
	os.WriteFile(dir+"/untracked.txt", []byte("x"), 0o644)

	g := New(dir)
	s, err := g.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if s.Staged != 1 {
		t.Fatalf("staged=%d, want 1", s.Staged)
	}
	if s.Uncommitted != 2 { // 1 modifiziert + 1 untracked
		t.Fatalf("uncommitted=%d, want 2", s.Uncommitted)
	}
}

func TestGitStatusDetached(t *testing.T) {
	dir := initRepo(t)
	commit(t, dir, "a.txt", "a")
	cmd := exec.Command("git", "-C", dir, "checkout", "-q", "HEAD~0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("checkout: %v (%s)", err, out)
	}
	g := New(dir)
	s, err := g.Read()
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if s.Branch == "" {
		t.Fatal("branch sollte bei detached HEAD der kurze Hash sein")
	}
}

func TestCountPorcelain(t *testing.T) {
	out := "M  a.txt\n M b.txt\nAM c.txt\n?? untracked.txt\n"
	staged, uncommitted := countPorcelain(out)
	if staged != 2 { // M + AM (Index)
		t.Fatalf("staged=%d, want 2", staged)
	}
	if uncommitted != 3 { // M (b), AM (worktree), ?? untracked
		t.Fatalf("uncommitted=%d, want 3", uncommitted)
	}
}
