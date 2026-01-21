package capture

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sort"
	"strings"
	"time"
)

// CaptureHeap writes a heap profile into dir using a timestamped filename.
// It returns the full file path.
func CaptureHeap(dir, prefix string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "./profiles"
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("capture: mkdir %s: %w", dir, err)
	}
	if prefix == "" {
		prefix = "heap"
	}
	name := fmt.Sprintf("%s-%s.pb.gz", prefix, time.Now().UTC().Format("20060102-150405Z"))
	path := filepath.Join(dir, name)

	f, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("capture: create %s: %w", path, err)
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("capture: write heap profile: %w", err)
	}
	return path, nil
}

// Rotate keeps only the most recent 'maxFiles' files in dir that match the
// given prefix (if non-empty). Older files are deleted. If maxFiles <= 0, it is a no-op.
func Rotate(dir string, maxFiles int, prefix string) error {
	if maxFiles <= 0 {
		return nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // ignore errors silently for rotation
	}
	type fileInfo struct{ name string; mod time.Time }
	files := make([]fileInfo, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() { continue }
		name := e.Name()
		if prefix != "" && !strings.HasPrefix(name, prefix+"-") {
			continue
		}
		info, err := e.Info()
		if err != nil { continue }
		files = append(files, fileInfo{name: name, mod: info.ModTime()})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod.After(files[j].mod) })
	if len(files) <= maxFiles {
		return nil
	}
	for _, f := range files[maxFiles:] {
		_ = os.Remove(filepath.Join(dir, f.name))
	}
	return nil
}
