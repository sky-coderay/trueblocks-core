package statePkg

import (
	"context"
	"fmt"
	"sort"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/call"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/identifiers"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/output"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/utils"
)

func (opts *StateOptions) HandleCall() error {
	chain := opts.Globals.Chain
	testMode := opts.Globals.TestMode
	isApiMode := opts.Globals.IsApiMode()

	callAddress := base.HexToAddress(opts.Addrs[0])
	if opts.ProxyFor != "" {
		callAddress = base.HexToAddress(opts.ProxyFor)
	}

	contractCall, err := call.NewContractCall(opts.Conn, callAddress, opts.Call, !isApiMode)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	fetchData := func(modelChan chan types.Modeler[types.RawResult], errorChan chan error) {
		var err error
		var txMap map[identifiers.ResolvedId]*types.SimpleResult
		if txMap, _, err = identifiers.AsMap[types.SimpleResult](chain, opts.BlockIds); err != nil {
			errorChan <- err
			cancel()
		}

		bar := logger.NewBar("", !opts.Globals.TestMode && len(opts.Globals.File) == 0, int64(len(txMap)))

		iterCtx, iterCancel := context.WithCancel(context.Background())
		defer iterCancel()

		nErrors := 0
		iterFunc := func(app identifiers.ResolvedId, value *types.SimpleResult) error {
			contractCall.BlockNumber = app.BlockNumber
			results, err := contractCall.Call12()
			if err != nil {
				errorChan <- err
				cancel()
			} else {
				if testMode {
					msg := fmt.Sprintf("call to %s at block %d at four-byte %s returned %v",
						contractCall.Address.Hex(), contractCall.BlockNumber, contractCall.Method.Encoding, results.Outputs)
					logger.TestLog(true, msg)
				}
				bar.Tick()
				*value = *results
			}
			return nil
		}

		iterErrorChan := make(chan error)
		go utils.IterateOverMap(iterCtx, iterErrorChan, txMap, iterFunc)
		for err := range iterErrorChan {
			// TODO: I don't really want to quit looping here. Just report the error and keep going.
			// iterCancel()
			if !opts.Globals.TestMode || nErrors == 0 {
				errorChan <- err
				// Reporting more than one error causes tests to fail because they
				// appear concurrently so sort differently
				nErrors++
			}
		}
		bar.Finish(true)

		items := make([]types.SimpleResult, 0, len(txMap))
		for _, v := range txMap {
			v := v
			items = append(items, *v)
		}
		sort.Slice(items, func(i, j int) bool {
			return items[i].BlockNumber < items[j].BlockNumber
		})

		for _, item := range items {
			item := item
			modelChan <- &item
		}

	}

	return output.StreamMany(ctx, fetchData, opts.Globals.OutputOptsWithExtra(nil))
}
