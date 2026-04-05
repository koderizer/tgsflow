package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/kelvin/tgsflow/src/templates"
	"github.com/kelvin/tgsflow/src/util/logx"
	"github.com/spf13/cobra"
)

// CmdInit implements `tgs init [--force]`.
//
// It copies the embedded scaffold tree (tgs/, .claude/, CLAUDE.md,
// Makefile.tgs.mk) into the current directory. Existing files are preserved
// by default; --force overwrites them (tgs/thoughts/ is always preserved).
// After copying, it ensures the repo Makefile includes Makefile.tgs.mk so
// `make new-thought` is available immediately.
func CmdInit(args []string) int {
	fs := flag.NewFlagSet("tgs init", flag.ContinueOnError)
	force := fs.Bool("force", false, "Overwrite existing scaffold files (tgs/thoughts/ is always preserved)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if len(fs.Args()) > 0 {
		fmt.Fprintln(os.Stderr, "usage: tgs init [--force]")
		return 2
	}

	logx.Infof("initializing TGS scaffold (force=%v)", *force)
	if err := templates.CopyScaffold(".", *force); err != nil {
		logx.Errorf("copy scaffold: %v", err)
		return 1
	}

	if err := ensureMakefileInclude(); err != nil {
		logx.Errorf("ensure Makefile include: %v", err)
		return 1
	}

	logx.Infof("tgs init complete (idempotent)")
	return 0
}

// Cobra command constructor.
func newInitCommand() *cobra.Command {
	var flagForce bool
	cmd := &cobra.Command{
		Use:     "init",
		Short:   "Copy the TGS scaffold into the current directory (idempotent)",
		Long:    "Copy the embedded TGS scaffold (tgs/, .claude/, CLAUDE.md, Makefile.tgs.mk) into the current directory. Existing files are preserved by default. Use --force to overwrite (files under tgs/thoughts/ are always preserved).",
		Example: "  tgs init\n  tgs init --force",
		Args:    cobra.NoArgs,
		RunE: func(c *cobra.Command, args []string) error {
			forward := []string{}
			if c.Flags().Changed("force") && flagForce {
				forward = append(forward, "--force")
			}
			return codeToErr(CmdInit(forward))
		},
	}
	cmd.Flags().BoolVar(&flagForce, "force", false, "Overwrite existing scaffold files (tgs/thoughts/ is always preserved)")
	return cmd
}

// ensureMakefileInclude appends `include Makefile.tgs.mk` to the repo Makefile
// if no new-thought target is defined and the include line is missing.
// If Makefile does not exist, it is created with the include line.
func ensureMakefileInclude() error {
	const includeLine = "include Makefile.tgs.mk"
	const targetHeader = "new-thought:"
	path := "Makefile"

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return os.WriteFile(path, []byte(includeLine+"\n"), 0o644)
		}
		return err
	}
	s := string(data)
	if strings.Contains(s, includeLine) || strings.Contains(s, targetHeader) {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\n" + includeLine + "\n")
	return err
}
