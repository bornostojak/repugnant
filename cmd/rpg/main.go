package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

const version = "0.1.0-dev"

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "rpg:", err)
		os.Exit(1)
	}
}

func run(args []string, stdout, _ io.Writer) error {
	flags := flag.NewFlagSet("rpg", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	showVersion := flags.Bool("version", false, "print version")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(stdout, version)
		return err
	}
	_, err := fmt.Fprintln(stdout, "rpg: documentation that stays close to code")
	return err
}
