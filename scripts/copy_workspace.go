package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	if err := run(); err != nil {
		fatalf("copy workspace failed: %v", err)
	}
}

func run() error {
	src := filepath.Clean("workspace")
	dst := filepath.Clean("cmd/picoclaw/internal/onboard/workspace")

	if err := os.RemoveAll(dst); err != nil {
		return fmt.Errorf("remove destination %q: %w", dst, err)
	}

	if err := copyDir(src, dst); err != nil {
		return err
	}

	return nil
}

func copyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source %q: %w", src, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source %q is not a directory", src)
	}

	if err := os.MkdirAll(dst, info.Mode().Perm()); err != nil {
		return fmt.Errorf("create destination %q: %w", dst, err)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read directory %q: %w", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
			continue
		}

		if err := copyFile(srcPath, dstPath); err != nil {
			return err
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file %q: %w", src, err)
	}
	defer srcFile.Close()

	info, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("stat source file %q: %w", src, err)
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode().Perm())
	if err != nil {
		return fmt.Errorf("create destination file %q: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("copy file %q -> %q: %w", src, dst, err)
	}

	return nil
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
