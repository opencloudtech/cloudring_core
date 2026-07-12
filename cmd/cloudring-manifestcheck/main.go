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
	report, err := platformmanifest.VerifySecretManager(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, "platform manifest verification blocked")
		os.Exit(1)
	}
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, "encode platform manifest report")
		os.Exit(2)
	}
}
