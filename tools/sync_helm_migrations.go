package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

func main() {
	root, err := resolveRoot()
	if err != nil {
		fail("resolve repository root", err)
	}

	sourceDir := filepath.Join(root, "db", "migrations")
	targetDir := filepath.Join(root, "deployment", "helm", "migrations")

	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		fail("read source migrations", err)
	}

	migrationNames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		migrationNames = append(migrationNames, entry.Name())
	}

	sort.Strings(migrationNames)

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		fail("create target migrations directory", err)
	}

	existingEntries, err := os.ReadDir(targetDir)
	if err != nil {
		fail("read target migrations", err)
	}

	for _, entry := range existingEntries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		if err := os.Remove(filepath.Join(targetDir, entry.Name())); err != nil {
			fail("remove stale target migration", err)
		}
	}

	for _, name := range migrationNames {
		sourcePath := filepath.Join(sourceDir, name)
		targetPath := filepath.Join(targetDir, name)

		if err := copyFile(sourcePath, targetPath); err != nil {
			fail(fmt.Sprintf("copy %s", name), err)
		}
	}

	fmt.Printf("synced %d Flyway migrations into %s\n", len(migrationNames), targetDir)
}

func resolveRoot() (string, error) {
	if len(os.Args) > 1 {
		return filepath.Abs(os.Args[1])
	}

	workingDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	currentDir := workingDir
	for {
		if isRepoRoot(currentDir) {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", errors.New("could not find repository root containing db/migrations and deployment/helm")
		}
		currentDir = parentDir
	}
}

func isRepoRoot(path string) bool {
	sourceDir := filepath.Join(path, "db", "migrations")
	chartDir := filepath.Join(path, "deployment", "helm")

	sourceInfo, sourceErr := os.Stat(sourceDir)
	chartInfo, chartErr := os.Stat(chartDir)

	return sourceErr == nil && chartErr == nil && sourceInfo.IsDir() && chartInfo.IsDir()
}

func copyFile(sourcePath, targetPath string) error {
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(targetPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(targetFile, sourceFile); err != nil {
		targetFile.Close()
		return err
	}

	if err := targetFile.Close(); err != nil {
		return err
	}

	return os.Chmod(targetPath, 0o644)
}

func fail(action string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", action, err)
	os.Exit(1)
}
