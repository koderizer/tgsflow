// Package templates embeds the TGS scaffold and copies it into target repos.
// No text templating: the scaffold is a tree of plain files that are copied
// verbatim. Executable bit is applied to *.sh files.
package templates

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// scaffoldFS contains the embedded scaffold tree. The `all:` prefix includes
// dotfiles (.claude/, .gitkeep) which are otherwise excluded by go:embed.
//
//go:embed all:scaffold
var scaffoldFS embed.FS

// ScaffoldFS returns the embedded scaffold filesystem. Exposed for tests.
func ScaffoldFS() fs.FS { return scaffoldFS }

// CopyScaffold walks the embedded scaffold tree and copies each file into
// destRoot. Existing files are preserved by default (idempotent). When force
// is true, existing files are overwritten — EXCEPT files under tgs/thoughts/
// which are always preserved because that directory holds user content.
func CopyScaffold(destRoot string, force bool) error {
	return fs.WalkDir(scaffoldFS, "scaffold", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == "scaffold" {
			return nil
		}
		rel := strings.TrimPrefix(path, "scaffold/")
		out := filepath.Join(destRoot, filepath.FromSlash(rel))

		if d.IsDir() {
			return os.MkdirAll(out, 0o755)
		}

		// Never overwrite files under tgs/thoughts/ (user content).
		underThoughts := strings.HasPrefix(rel, "tgs/thoughts/")
		if _, statErr := os.Stat(out); statErr == nil {
			if underThoughts || !force {
				return nil
			}
		}

		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		content, rerr := scaffoldFS.ReadFile(path)
		if rerr != nil {
			return rerr
		}
		mode := os.FileMode(0o644)
		if strings.HasSuffix(rel, ".sh") {
			mode = 0o755
		}
		return os.WriteFile(out, content, mode)
	})
}
