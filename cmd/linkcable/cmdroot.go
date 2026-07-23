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
	"errors"
	"log/slog"
	"net/http"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/log"
	"github.com/thediveo/mobydriver/network"

	driver "github.com/thediveo/linkcable"
)

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "linkcable",
		Short:   "linkcable is a Docker/Moby network driver plugin for point-to-point networks",
		Version: version(),
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: runLinkcable,
	}
	clippy.AddFlags(rootCmd)
	return rootCmd
}

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	v := info.Main.Version
	if v == "" {
		return "unknown"
	}
	return v
}

func runLinkcable(cmd *cobra.Command, _ []string) error {
	slog.Info("linkcable Docker network driver plugin",
		slog.String("version", cmd.Version),
		slog.String("license", "Apache 2.0"),
		slog.String("web", "https://github.com/thediveo/linkcable"))

	lcdriver, err := driver.NewDriver("linkcable")
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	h := network.NewHandler(lcdriver)
	err = h.ServeUnix(ctx, "linkcable", 0) // FIXME: name, gid
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
