// Copyright 2021 The TrueBlocks Authors. All rights reserved.
// Use of this source code is governed by a license that can
// be found in the LICENSE file.
/*
 * Parts of this file were generated with makeClass --run. Edit only those parts of
 * the code inside of 'EXISTING_CODE' tags.
 */

package slurpPkg

// EXISTING_CODE
import (
	"net/http"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/internal/globals"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/output"
	outputHelpers "github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/output/helpers"
	"github.com/spf13/cobra"
)

// EXISTING_CODE

// RunSlurp handles the slurp command for the command line. Returns error only as per cobra.
func RunSlurp(cmd *cobra.Command, args []string) (err error) {
	opts := slurpFinishParse(args)
	// EXISTING_CODE
	// EXISTING_CODE
	err, _ = opts.SlurpInternal()
	outputHelpers.CloseJsonWriterIfNeeded(func() *globals.GlobalOptions {
		return &opts.Globals
	})
	return
}

// ServeSlurp handles the slurp command for the API. Returns error and a bool if handled
func ServeSlurp(w http.ResponseWriter, r *http.Request) (err error, handled bool) {
	opts := slurpFinishParseApi(w, r)
	// EXISTING_CODE
	// EXISTING_CODE
	err, handled = opts.SlurpInternal()
	if opts.Globals.Format == "json" && err == nil {
		opts.Globals.Writer.(*output.JsonWriter).Close()
	}
	return
}

// SlurpInternal handles the internal workings of the slurp command.  Returns error and a bool if handled
func (opts *SlurpOptions) SlurpInternal() (err error, handled bool) {
	err = opts.validateSlurp()
	if err != nil {
		return err, true
	}

	// EXISTING_CODE
	if opts.Globals.IsApiMode() {
		return nil, false
	}

	handled = true
	err = opts.Globals.PassItOn("ethslurp", opts.Globals.Chain, opts.toCmdLine(), opts.getEnvStr())
	// EXISTING_CODE

	return
}

// EXISTING_CODE
// EXISTING_CODE
