// Copyright (c) Facebook, Inc. and its affiliates.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This is a generated file, edits should be made in the corresponding source
// file, and this file regenerated using
//   `contest-generator --from core-plugins.yml`
// followed by
//   `gofmt -w contest.go`

package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/linuxboot/contest/cmds/contest/server"

	// the targetmanager plugins
	csvtargetmanager "github.com/linuxboot/contest/plugins/targetmanagers/csvtargetmanager"
	targetlist "github.com/linuxboot/contest/plugins/targetmanagers/targetlist"

	// the testfetcher plugins
	literal "github.com/linuxboot/contest/plugins/testfetchers/literal"
	uri "github.com/linuxboot/contest/plugins/testfetchers/uri"

	// the teststep plugins
	ts_cmd "github.com/linuxboot/contest/plugins/teststeps/cmd"
	cpucmd "github.com/linuxboot/contest/plugins/teststeps/cpucmd"
	echo "github.com/linuxboot/contest/plugins/teststeps/echo"
	exec "github.com/linuxboot/contest/plugins/teststeps/exec"
	gathercmd "github.com/linuxboot/contest/plugins/teststeps/gathercmd"
	randecho "github.com/linuxboot/contest/plugins/teststeps/randecho"
	sleep "github.com/linuxboot/contest/plugins/teststeps/sleep"
	sshcmd "github.com/linuxboot/contest/plugins/teststeps/sshcmd"

	// the reporter plugins
	noop "github.com/linuxboot/contest/plugins/reporters/noop"
	targetsuccess "github.com/linuxboot/contest/plugins/reporters/targetsuccess"
)

func getPluginConfig() *server.PluginConfig {
	var pc server.PluginConfig
	pc.TargetManagerLoaders = append(pc.TargetManagerLoaders, csvtargetmanager.Load)
	pc.TargetManagerLoaders = append(pc.TargetManagerLoaders, targetlist.Load)
	pc.TestFetcherLoaders = append(pc.TestFetcherLoaders, literal.Load)
	pc.TestFetcherLoaders = append(pc.TestFetcherLoaders, uri.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, ts_cmd.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, cpucmd.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, echo.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, exec.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, gathercmd.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, randecho.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, sleep.Load)
	pc.TestStepLoaders = append(pc.TestStepLoaders, sshcmd.Load)
	pc.ReporterLoaders = append(pc.ReporterLoaders, noop.Load)
	pc.ReporterLoaders = append(pc.ReporterLoaders, targetsuccess.Load)

	return &pc
}

func main() {
	sigs := make(chan os.Signal, 1)
	defer close(sigs)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
	if err := server.Main(getPluginConfig(), os.Args[0], os.Args[1:], sigs); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
