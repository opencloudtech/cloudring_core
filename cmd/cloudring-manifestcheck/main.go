// SPDX-License-Identifier: Apache-2.0
// Copyright (C) IURII TRUKHIN 2012-2022, Elena Trukhina 2023-2026. Project and trademarks: Elena Trukhina ZZP.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/opencloudtech/CloudRING/internal/platformmanifest"
)

func main() {
	root := flag.String("root", ".", "repository root")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(2)
	}
	reports := make([]platformmanifest.Report, 0, 7)
	for _, verify := range []func(string) (platformmanifest.Report, error){
		platformmanifest.VerifySecretManager,
		platformmanifest.VerifyCertManager,
		platformmanifest.VerifyCSISnapshotAPI,
		platformmanifest.VerifyRookCephRBD,
		platformmanifest.VerifyLonghornThreeNode,
		platformmanifest.VerifyPostgreSQLHA,
		platformmanifest.VerifyCDI,
	} {
		report, err := verify(*root)
		if err != nil {
			fmt.Fprintln(os.Stderr, "platform manifest verification blocked")
			os.Exit(1)
		}
		reports = append(reports, report)
	}
	output := struct {
		Status   string                    `json:"status"`
		Profiles []platformmanifest.Report `json:"profiles"`
	}{
		Status:   "ready",
		Profiles: reports,
	}
	if err := json.NewEncoder(os.Stdout).Encode(output); err != nil {
		fmt.Fprintln(os.Stderr, "encode platform manifest report")
		os.Exit(2)
	}
}
