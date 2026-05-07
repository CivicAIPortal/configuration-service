package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadConfigValidatesRequiredFields(t *testing.T) {
	repoRoot := t.TempDir()
	configDir := filepath.Join(repoRoot, ".ai")
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}

	content := "ai_config_repo: ../ai-governance\nversion: v0.1.0\nagent: backend-agent\noverrides: []\n"
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	config, err := LoadConfig(repoRoot)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if config.Agent != "backend-agent" {
		t.Fatalf("unexpected agent: %s", config.Agent)
	}
}

func TestLoadAgentManifestMatchesByName(t *testing.T) {
	governanceDir := t.TempDir()
	agentsDir := filepath.Join(governanceDir, "agents")
	if err := os.MkdirAll(agentsDir, 0o755); err != nil {
		t.Fatalf("mkdir agents dir: %v", err)
	}

	manifest := "name: backend-agent\nincludes:\n  - policies/coding.md\n"
	if err := os.WriteFile(filepath.Join(agentsDir, "backend.yaml"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	path, loaded, err := LoadAgentManifest(governanceDir, "backend-agent")
	if err != nil {
		t.Fatalf("LoadAgentManifest returned error: %v", err)
	}

	if !strings.HasSuffix(path, filepath.Join("agents", "backend.yaml")) {
		t.Fatalf("unexpected path: %s", path)
	}
	if len(loaded.Includes) != 1 || loaded.Includes[0] != "policies/coding.md" {
		t.Fatalf("unexpected includes: %#v", loaded.Includes)
	}
}

func TestResolveOverrideFilesRejectsMissingFiles(t *testing.T) {
	_, err := resolveOverrideFiles(t.TempDir(), []string{"missing.md"})
	if err == nil {
		t.Fatal("expected missing override to fail validation")
	}
}


func makeGitRepo(t *testing.T, dir string) {
t.Helper()
for _, c := range [][]string{
{"git", "init", "-b", "main"},
{"git", "config", "user.email", "test@test.com"},
{"git", "config", "user.name", "Test"},
} {
cmd := exec.Command(c[0], c[1:]...)
cmd.Dir = dir
if out, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("setup %v: %v\n%s", c, err, out)
}
}
}

func addCommit(t *testing.T, dir, filename, content string) string {
t.Helper()
if err := os.WriteFile(filepath.Join(dir, filename), []byte(content), 0o644); err != nil {
t.Fatalf("write file: %v", err)
}
for _, c := range [][]string{
{"git", "add", filename},
{"git", "commit", "-m", "add " + filename},
} {
cmd := exec.Command(c[0], c[1:]...)
cmd.Dir = dir
if out, err := cmd.CombinedOutput(); err != nil {
t.Fatalf("%v: %v\n%s", c, err, out)
}
}
cmd := exec.Command("git", "rev-parse", "HEAD")
cmd.Dir = dir
out, _ := cmd.CombinedOutput()
return strings.TrimSpace(string(out))
}

func TestResolveGovernanceRepoAdvancesCacheAfterNewCommit(t *testing.T) {
originDir := t.TempDir()
makeGitRepo(t, originDir)
if err := os.MkdirAll(filepath.Join(originDir, "agents"), 0o755); err != nil {
t.Fatalf("mkdir agents: %v", err)
}
addCommit(t, originDir, filepath.Join("agents", "backend.yaml"), "name: backend-agent\nincludes: []\n")

repoRoot := t.TempDir()
cacheRoot := t.TempDir()
t.Setenv("HOME", cacheRoot)
t.Setenv("LOCALAPPDATA", cacheRoot)
config := Config{AIConfigRepo: originDir, Version: "main", Agent: "backend-agent"}

dir1, err := ResolveGovernanceRepo(repoRoot, config)
if err != nil {
t.Fatalf("first resolve: %v", err)
}
rev1, _ := gitRevision(dir1)

rev2 := addCommit(t, originDir, "extra.md", "new content\n")
if rev1 == rev2 {
t.Fatal("test setup: rev1 and rev2 should differ")
}

dir2, err := ResolveGovernanceRepo(repoRoot, config)
if err != nil {
t.Fatalf("second resolve: %v", err)
}
got, _ := gitRevision(dir2)
if got != rev2 {
t.Fatalf("cache not advanced: want %s, got %s", rev2, got)
}
}
