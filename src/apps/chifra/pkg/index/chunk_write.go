package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/cache"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/index/bloom"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/version"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type AddressAppearanceMap map[string][]AppearanceRecord

type WriteChunkReport struct {
	Range        cache.FileRange
	nAddresses   int
	nAppearances int
	Snapped      bool
	Pinned       bool
}

func (c *WriteChunkReport) Report() {
	str := fmt.Sprintf("%sWrote %d address and %d appearance records to $INDEX/%s.bin%s%s", colors.BrightBlue, c.nAddresses, c.nAppearances, c.Range, colors.Off, spaces20)
	logger.Log(logger.Info, str)
	if c.Pinned {
		str := fmt.Sprintf("%sPinned chunk $INDEX/%s.bin%s%s", colors.BrightBlue, c.Range, colors.Off, spaces20)
		logger.Log(logger.Info, str)
	}
}

func WriteChunk(chain, fileName string, addrAppearanceMap AddressAppearanceMap, nApps int) (*WriteChunkReport, error) {
	// We're going to build two tables. An addressTable and an appearanceTable. We do this as we spin
	// through the map

	// Create space for the two tables...
	addressTable := make([]AddressRecord, 0, len(addrAppearanceMap))
	appearanceTable := make([]AppearanceRecord, 0, nApps)

	// We want to sort the items in the map by address (maps in GoLang are not sorted)
	sorted := []string{}
	for address := range addrAppearanceMap {
		sorted = append(sorted, address)
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	// We need somewhere to store our progress...
	offset := uint32(0)
	bl := bloom.ChunkBloom{}

	// For each address in the sorted list...
	for _, addrStr := range sorted {
		// ...get its appeances and append them to the appearanceTable....
		apps := addrAppearanceMap[addrStr]
		appearanceTable = append(appearanceTable, apps...)

		// ...add the address to the bloom filter...
		address := common.HexToAddress(addrStr)
		bl.AddToSet(address)

		// ...and append the record to the addressTable.
		addressTable = append(addressTable, AddressRecord{
			Address: address,
			Offset:  offset,
			Count:   uint32(len(apps)),
		})

		// Finally, note the next offset
		offset += uint32(len(apps))
	}

	// At this point, the two tables and the bloom filter are fully populated. We're ready to write to disc...

	// First, we backup the existing chunk if there is one...
	indexFn := ToIndexPath(fileName)
	tmpPath := filepath.Join(config.GetPathToCache(chain), "tmp")
	if backupFn, err := file.MakeBackup(tmpPath, indexFn); err == nil {
		defer func() {
			if file.FileExists(backupFn) {
				// If the backup file exists, something failed, so we replace the original file.
				os.Rename(backupFn, indexFn)
				os.Remove(backupFn) // seems redundant, but may not be on some operating systems
			}
		}()

		if fp, err := os.OpenFile(indexFn, os.O_WRONLY|os.O_CREATE, 0644); err == nil {
			defer fp.Close() // defers are last in, first out

			fp.Seek(0, io.SeekStart) // already true, but can't hurt
			header := HeaderRecord{
				Magic:           file.MagicNumber,
				Hash:            common.BytesToHash(crypto.Keccak256([]byte(version.ManifestVersion))),
				AddressCount:    uint32(len(addressTable)),
				AppearanceCount: uint32(len(appearanceTable)),
			}
			if err = binary.Write(fp, binary.LittleEndian, header); err != nil {
				return nil, err
			}

			if binary.Write(fp, binary.LittleEndian, addressTable); err != nil {
				return nil, err
			}

			if binary.Write(fp, binary.LittleEndian, appearanceTable); err != nil {
				return nil, err
			}

			if _, err = bl.WriteBloom(chain, bloom.ToBloomPath(indexFn)); err != nil {
				return nil, err
			}

			// Success. Remove the backup so it doesn't replace the orignal
			os.Remove(backupFn)
			rng, _ := cache.RangeFromFilename(indexFn)
			return &WriteChunkReport{
				Range:        rng,
				nAddresses:   len(addressTable),
				nAppearances: len(appearanceTable),
			}, nil

		} else {
			return nil, err
		}

	} else {
		return nil, err
	}
}

type Renderer interface {
	RenderObject(data interface{}, first bool) error
}

var spaces20 = strings.Repeat(" ", 20)
