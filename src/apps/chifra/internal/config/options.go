// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.
/*
 * This file was auto generated with makeClass --gocmds. DO NOT EDIT.
 */

package configPkg

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/internal/globals"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/validate"
)

// ConfigOptions provides all command options for the chifra config command.
type ConfigOptions struct {
	Modes   []string              `json:"modes,omitempty"`   // Either show or edit the configuration
	Module  []string              `json:"module,omitempty"`  // The type of information to show or edit
	Types   []string              `json:"types,omitempty"`   // For caches module only, which type(s) of cache to report
	Paths   bool                  `json:"paths,omitempty"`   // Show the configuration paths for the system
	Globals globals.GlobalOptions `json:"globals,omitempty"` // The global options
	BadFlag error                 `json:"badFlag,omitempty"` // An error flag if needed
	// EXISTING_CODE
	// EXISTING_CODE
}

var defaultConfigOptions = ConfigOptions{}

// testLog is used only during testing to export the options for this test case.
func (opts *ConfigOptions) testLog() {
	logger.TestLog(len(opts.Modes) > 0, "Modes: ", opts.Modes)
	logger.TestLog(len(opts.Module) > 0, "Module: ", opts.Module)
	logger.TestLog(len(opts.Types) > 0, "Types: ", opts.Types)
	logger.TestLog(opts.Paths, "Paths: ", opts.Paths)
	opts.Globals.TestLog()
}

// String implements the Stringer interface
func (opts *ConfigOptions) String() string {
	b, _ := json.MarshalIndent(opts, "", "  ")
	return string(b)
}

// getEnvStr allows for custom environment strings when calling to the system (helps debugging).
func (opts *ConfigOptions) getEnvStr() []string {
	envStr := []string{}
	// EXISTING_CODE
	// EXISTING_CODE
	return envStr
}

// toCmdLine converts the option to a command line for calling out to the system.
func (opts *ConfigOptions) toCmdLine() string {
	options := ""
	for _, module := range opts.Module {
		options += " --module " + module
	}
	for _, types := range opts.Types {
		options += " --types " + types
	}
	options += " " + strings.Join(opts.Modes, " ")
	// EXISTING_CODE
	// EXISTING_CODE
	options += fmt.Sprintf("%s", "") // silence compiler warning for auto gen
	return options
}

// configFinishParseApi finishes the parsing for server invocations. Returns a new ConfigOptions.
func configFinishParseApi(w http.ResponseWriter, r *http.Request) *ConfigOptions {
	copy := defaultConfigOptions
	opts := &copy
	for key, value := range r.URL.Query() {
		switch key {
		case "modes":
			for _, val := range value {
				s := strings.Split(val, " ") // may contain space separated items
				opts.Modes = append(opts.Modes, s...)
			}
		case "module":
			for _, val := range value {
				s := strings.Split(val, " ") // may contain space separated items
				opts.Module = append(opts.Module, s...)
			}
		case "types":
			for _, val := range value {
				s := strings.Split(val, " ") // may contain space separated items
				opts.Types = append(opts.Types, s...)
			}
		case "paths":
			opts.Paths = true
		default:
			if !globals.IsGlobalOption(key) {
				opts.BadFlag = validate.Usage("Invalid key ({0}) in {1} route.", key, "config")
				return opts
			}
		}
	}
	opts.Globals = *globals.GlobalsFinishParseApi(w, r)
	// EXISTING_CODE
	// EXISTING_CODE

	return opts
}

// configFinishParse finishes the parsing for command line invocations. Returns a new ConfigOptions.
func configFinishParse(args []string) *ConfigOptions {
	opts := GetOptions()
	opts.Globals.FinishParse(args)
	defFmt := "txt"
	// EXISTING_CODE
	defFmt = ""
	for _, mode := range args {
		if mode == "show" || mode == "edit" {
			opts.Modes = append(opts.Modes, mode)
		} else {
			opts.Module = append(opts.Module, mode)
		}
	}
	if len(opts.Modes) == 0 {
		opts.Modes = []string{"show"}
	}
	// EXISTING_CODE
	if len(opts.Globals.Format) == 0 || opts.Globals.Format == "none" {
		opts.Globals.Format = defFmt
	}
	return opts
}

func GetOptions() *ConfigOptions {
	// EXISTING_CODE
	// EXISTING_CODE
	return &defaultConfigOptions
}

func ResetOptions() {
	// We want to keep writer between command file calls
	w := GetOptions().Globals.Writer
	defaultConfigOptions = ConfigOptions{}
	globals.SetDefaults(&defaultConfigOptions.Globals)
	defaultConfigOptions.Globals.Writer = w
}
