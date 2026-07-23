// Copyright 2026 Harald Albrecht.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/log"
	"github.com/thediveo/mobydriver/network"

	driver "github.com/thediveo/linkcable"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "linkcable",
		Short: "linkcable is a Docker/Moby network driver plugin for point-to-point networks",
		Args:  cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: runLinkcable,
	}
	clippy.AddFlags(rootCmd)
	return rootCmd
}

func runLinkcable(cmd *cobra.Command, _ []string) error {
	slog.Info("linkcable Docker network driver plugin",
		slog.String("license", "Apache 2.0"),
		slog.String("web", "https://github.com/thediveo/linkcable")) // TODO: version

	lcdriver, err := driver.NewDriver("linkcable")
	if err != nil {
		return err
	}
	h := network.NewHandler(lcdriver)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	return h.ServeUnix(ctx, "linkcable", 0) // FIXME: name, gid
}
