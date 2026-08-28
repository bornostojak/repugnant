package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bornostojak/repugnant/internal/docs"
	"github.com/bornostojak/repugnant/internal/project"
)

// version is overridable at build time via -ldflags "-X main.version=...".
var version = "0.1.0-dev"

const usageText = `rpg — documentation that stays close to code

Usage:
  rpg <command> [flags]

Commands:
  init                    Write rpg.conf.yaml, create .rpg/, install Git hooks
  generate                Parse sources and write/update local Markdown docs
  push                    Publish generated docs to the configured project API
  status                  Show whether a previous publish is pending
  project create          Create a project on a server and print config guidance
  hook <pre-commit|pre-push>  Run an installed Git hook (invoked by Git)
  help                    Show this help

Flags:
  --init-hooks            Install or repair only the Git hooks
  --version               Print the CLI version
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "rpg:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return usage(stdout)
	}
	switch args[0] {
	case "help", "-h", "--help":
		return usage(stdout)
	case "init":
		return initialize(false, stdout)
	case "generate":
		return generate(args[1:], stdout)
	case "push":
		return push(stdout)
	case "status":
		return status(stdout)
	case "project":
		if len(args) > 1 && args[1] == "create" {
			return createProject(args[2:], stdout)
		}
		return fmt.Errorf("usage: rpg project create --server http://host:8080 --name 'My project' [--slug my-project]")
	case "hook":
		return runHook(args[1:], stdout, stderr)
	}
	if !strings.HasPrefix(args[0], "-") {
		return fmt.Errorf("unknown command %q; run 'rpg help' for usage", args[0])
	}
	flags := flag.NewFlagSet("rpg", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	showVersion := flags.Bool("version", false, "print version")
	initHooks := flags.Bool("init-hooks", false, "install or repair hooks only")
	if err := flags.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return usage(stdout)
		}
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}
	if *initHooks {
		return initialize(true, stdout)
	}
	return usage(stdout)
}

func usage(stdout io.Writer) error {
	_, err := fmt.Fprint(stdout, usageText)
	return err
}

func status(stdout io.Writer) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	rec, err := docs.LoadPending(root)
	if err != nil {
		return err
	}
	if rec == nil {
		_, err = fmt.Fprintln(stdout, "rpg status: no pending documentation")
		return err
	}
	_, err = fmt.Fprintf(stdout, "rpg status: publish pending since %s\n  error: %s\n  run 'rpg push' to retry\n", rec.FailedAt, rec.Error)
	return err
}
func createProject(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("rpg project create", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	server := flags.String("server", "", "server base URL")
	name := flags.String("name", "", "project name")
	slug := flags.String("slug", "", "project slug")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *server == "" || *name == "" {
		return fmt.Errorf("usage: rpg project create --server http://host:8080 --name 'My project' [--slug my-project]")
	}
	body, _ := json.Marshal(map[string]string{"name": *name, "slug": *slug})
	resp, err := http.Post(strings.TrimRight(*server, "/")+"/api/projects", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("create project: server returned %s", resp.Status)
	}
	var result struct {
		Slug   string `json:"slug"`
		APIKey string `json:"api_key"`
		APIURL string `json:"api_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("read project response: %w", err)
	}
	apiURL := result.APIURL
	if strings.HasPrefix(apiURL, "/") {
		apiURL = strings.TrimRight(*server, "/") + apiURL
	}
	_, err = fmt.Fprintf(stdout, "project %q created. Add this to rpg.conf.yaml (keep api_key private):\n\noutput:\n  web:\n    enabled: true\n    endpoint: %s\nproject:\n  slug: %s\n  api_url: %s\n  api_key: %s\n", result.Slug, strings.TrimRight(*server, "/"), result.Slug, apiURL, result.APIKey)
	return err
}
func push(stdout io.Writer) error {
	root, e := os.Getwd()
	if e != nil {
		return e
	}
	n, e := docs.Push(root)
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(stdout, "published %d documentation article(s)\n", n)
	return e
}
func generate(args []string, stdout io.Writer) error {
	flags := flag.NewFlagSet("rpg generate", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	staged := flags.Bool("staged", false, "only process files staged in git")
	if err := flags.Parse(args); err != nil {
		return err
	}
	root, e := os.Getwd()
	if e != nil {
		return e
	}
	res, e := docs.GenerateWith(root, docs.GenerateOptions{Staged: *staged})
	if e != nil {
		return e
	}
	_, e = fmt.Fprintf(stdout, "generated %d documentation article(s)\n", res.Count)
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

func runHook(args []string, stdout, stderr io.Writer) error {
	if len(args) != 1 || (args[0] != "pre-commit" && args[0] != "pre-push") {
		return fmt.Errorf("usage: rpg hook <pre-commit|pre-push>")
	}
	if args[0] == "pre-commit" {
		if _, err := os.Stat(project.ConfigFileName); err == nil {
			root, err := os.Getwd()
			if err != nil {
				return err
			}
			// Only touch files that are part of this commit, and stage the
			// marker rewrites back so the committed source matches the docs.
			res, err := docs.GenerateWith(root, docs.GenerateOptions{Staged: true})
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(stdout, "generated %d documentation article(s)\n", res.Count); err != nil {
				return err
			}
			config, err := project.Load(".")
			if err != nil {
				return err
			}
			if !config.Output.Docs.Enabled {
				return nil
			}
			paths := append([]string{config.Output.Docs.Dir}, res.ChangedFiles...)
			cmd := exec.Command("git", append([]string{"add", "--"}, paths...)...)
			cmd.Stdout, cmd.Stderr = stdout, stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("stage generated docs: %w", err)
			}
			return nil
		}
	}
	if args[0] == "pre-push" {
		if _, err := os.Stat(project.ConfigFileName); os.IsNotExist(err) {
			return nil
		}
		config, err := project.Load(".")
		if err != nil {
			return err
		}
		if !config.Output.Web.Enabled {
			return nil
		}
		if _, err := docs.Push("."); err == nil {
			_, err = fmt.Fprintln(stdout, "rpg pre-push: published staged documentation")
			return err
		} else if config.Hooks.OnPublishFailure == "allow_pending" {
			if isInteractive() {
				fmt.Fprintf(stderr, "rpg pre-push: publishing failed: %v\nContinue push and keep .rpg/pending.json? [y/N] ", err)
				answer, _ := bufio.NewReader(os.Stdin).ReadString('\n')
				if strings.EqualFold(strings.TrimSpace(answer), "y") || strings.EqualFold(strings.TrimSpace(answer), "yes") {
					return docs.RecordPending(".", err)
				}
			}
			return docs.RecordPending(".", err)
		} else {
			return fmt.Errorf("rpg pre-push blocked: %w; fix the server/configuration or set hooks.on_publish_failure: allow_pending", err)
		}
	}
	_, err := fmt.Fprintf(stdout, "rpg %s: no pending documentation changes\n", args[0])
	return err
}

func isInteractive() bool {
	info, err := os.Stdin.Stat()
	return err == nil && (info.Mode()&os.ModeCharDevice) != 0
}
