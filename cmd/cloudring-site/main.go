// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/opencloudtech/CloudRING/pkg/siteprofile"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "preflight" && args[0] != "plan" {
		fmt.Fprintln(stderr, "usage: cloudring-site preflight|plan --profile <path|->")
		return 2
	}
	command := args[0]
	flags := flag.NewFlagSet(command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	profilePath := flags.String("profile", "-", "provider site profile path or - for stdin")
	if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
		return 2
	}
	reader, closeReader, err := openProfile(*profilePath, stdin)
	if err != nil {
		fmt.Fprintln(stderr, "provider site profile unavailable")
		return 1
	}
	defer closeReader()
	profile, err := siteprofile.Parse(reader)
	if err != nil {
		fmt.Fprintln(stderr, "provider site profile invalid")
		return 1
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetIndent("", "  ")
	if command == "preflight" {
		report := siteprofile.Validate(profile)
		if err := encoder.Encode(report); err != nil {
			fmt.Fprintln(stderr, "encode provider site preflight")
			return 2
		}
		if report.Status != "ready" {
			return 1
		}
		return 0
	}
	plan, err := siteprofile.BuildPlan(profile)
	if err != nil {
		fmt.Fprintln(stderr, "provider site plan blocked")
		return 1
	}
	if err := encoder.Encode(plan); err != nil {
		fmt.Fprintln(stderr, "encode provider site plan")
		return 2
	}
	return 0
}

func openProfile(path string, stdin io.Reader) (io.Reader, func(), error) {
	if strings.TrimSpace(path) == "-" {
		if stdin == nil {
			return nil, func() {}, siteprofile.ErrInvalidProfile
		}
		return stdin, func() {}, nil
	}
	// #nosec G304 G703 -- the local operator explicitly selects the bounded, strict profile input.
	file, err := os.Open(path)
	if err != nil {
		return nil, func() {}, err
	}
	return file, func() { _ = file.Close() }, nil
}
