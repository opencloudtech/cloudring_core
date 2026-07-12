// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/opencloudtech/CloudRING/pkg/openbaoauth"
)

const (
	exitPlanned = 0
	exitUsage   = 1
	exitBlocked = 2
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 2 || args[0] != "plan" || args[1] != "kubernetes-auth" {
		fmt.Fprintln(stderr, "usage: cloudring-openbao plan kubernetes-auth < contract.json")
		return exitUsage
	}
	report, err := openbaoauth.Evaluate(stdin)
	if err != nil {
		fmt.Fprintln(stderr, "OpenBao plan input unavailable")
		return exitUsage
	}
	encoder := json.NewEncoder(stdout)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(report); err != nil {
		fmt.Fprintln(stderr, "encode OpenBao plan report")
		return exitUsage
	}
	if report.Status != "planned" {
		fmt.Fprintln(stderr, "OpenBao Kubernetes auth plan blocked")
		return exitBlocked
	}
	return exitPlanned
}
