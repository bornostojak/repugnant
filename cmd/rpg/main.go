package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bornostojak/repugnant/internal/docs"
	"github.com/bornostojak/repugnant/internal/project"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "rpg:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, _ io.Writer) error {
	if len(args) > 0 && args[0] == "init" {
		return initialize(false, stdout)
	}
	if len(args) > 0 && args[0] == "generate" {
		return generate(stdout)
	}
	if len(args) > 0 && args[0] == "push" {
		return fmt.Errorf("rpg push is not available until web publishing is configured")
	}
	if len(args) > 0 && args[0] == "hook" {
		return runHook(args[1:], stdout)
	}
	flags := flag.NewFlagSet("rpg", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	showVersion := flags.Bool("version", false, "print version")
	initHooks := flags.Bool("init-hooks", false, "install or repair hooks only")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}
	if *initHooks {
		return initialize(true, stdout)
	}
	_, err := fmt.Fprintln(stdout, "rpg: documentation that stays close to code")
	return err
}
func generate(stdout io.Writer) error {
	root, e := os.Getwd()
	if e != nil {
		return e
	}
	n, e := docs.Generate(root)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(stdout, "generated %d documentation article(s)\n", n)
	return e
}

func initialize(hookOnly bool, stdout io.Writer) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	root, err = filepath.Abs(root)
	if err != nil {
		return err
	}
	if err := project.Init(root, hookOnly); err != nil {
		return err
	}
	if hookOnly {
		_, err = fmt.Fprintln(stdout, "rpg hooks installed")
	} else {
		_, err = fmt.Fprintln(stdout, "rpg initialized: rpg.conf.yaml, .rpg/, and Git hooks are ready")
	}
	return err
}

func runHook(args []string, stdout io.Writer) error {
	if len(args) != 1 || (args[0] != "pre-commit" && args[0] != "pre-push") {
		return fmt.Errorf("usage: rpg hook <pre-commit|pre-push>")
	}
	if args[0] == "pre-commit" {
		if _, err := os.Stat(project.ConfigFileName); err == nil {
			return generate(stdout)
		}
	}
	_, err := fmt.Fprintf(stdout, "rpg %s: no pending documentation changes\n", args[0])
	return err
}
