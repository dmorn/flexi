// SPDX-FileCopyrightText: 2020 jecoz
//
// SPDX-License-Identifier: BSD-3-Clause

// +build linux freebsd darwin netbsd openbsd dragonfly

package flexi

import (
	"fmt"
	"os"
	"os/exec"
)

func run(cmd string, args ...string) error {
	return exec.Command(cmd, args...).Run()
}

func mount(addr, mtpt string) error {
	if err := os.MkdirAll(mtpt, os.ModePerm); err != nil {
		return fmt.Errorf("mount: %w", err)
	}
	return run("9", "mount", addr, mtpt)
}

func umount(mtpt string) error {
	return run("9", "umount", mtpt)
}
