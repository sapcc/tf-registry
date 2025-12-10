// SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/sapcc/tf-registry/internal/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
}
