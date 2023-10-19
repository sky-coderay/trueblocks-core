package index

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/base"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/colors"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/config"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/logger"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/types"
)

type AppearanceMap map[string]types.SimpleAppearance
type writeChunkReport struct {
	Range        base.FileRange
	nAddresses   int
	nAppearances int
	FileSize     int64
	Snapped      bool
}

func (c *writeChunkReport) Report() {
	report := `Wrote {%d} address and {%d} appearance records to {$INDEX/%s.bin}`
	if c.Snapped {
		report += ` @(snapped to grid)}`
	}
	report += " (size: {%d} , span: {%d})"
	logger.Info(colors.ColoredWith(fmt.Sprintf(report, c.nAddresses, c.nAppearances, c.Range, c.FileSize, c.Range.Span()), colors.BrightBlue))
}

func (chunk *Chunk) Write(chain, newTag string, unused bool, publisher base.Address, fileName string, addrAppearanceMap map[string][]AppearanceRecord, nApps int) (*writeChunkReport, error) {
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
	bl := Bloom{}

	// For each address in the sorted list...
	for _, addrStr := range sorted {
		// ...get its appearances and append them to the appearanceTable....
		apps := addrAppearanceMap[addrStr]
		appearanceTable = append(appearanceTable, apps...)

		// ...add the address to the bloom filter...
		address := base.HexToAddress(addrStr)
		bl.InsertAddress(address)

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
	tmpPath := filepath.Join(config.PathToCache(chain), "tmp")
	if backupFn, err := file.MakeBackup(tmpPath, indexFn); err == nil {
		defer func() {
			if file.FileExists(backupFn) {
				// If the backup file exists, something failed, so we replace the original file.
				_ = os.Rename(backupFn, indexFn)
				_ = os.Remove(backupFn) // seems redundant, but may not be on some operating systems
			}
		}()

		if fp, err := os.OpenFile(indexFn, os.O_WRONLY|os.O_CREATE, 0644); err == nil {
			// defer fp.Close() // Note -- we don't defer because we want to close the file and possibly pin it below...

			_, _ = fp.Seek(0, io.SeekStart) // already true, but can't hurt
			header := indexHeader{
				Magic:           file.MagicNumber,
				Hash:            base.BytesToHash([]byte(newTag)),
				AddressCount:    uint32(len(addressTable)),
				AppearanceCount: uint32(len(appearanceTable)),
			}
			if err = binary.Write(fp, binary.LittleEndian, header); err != nil {
				return nil, err
			}

			if err = binary.Write(fp, binary.LittleEndian, addressTable); err != nil {
				return nil, err
			}

			if err = binary.Write(fp, binary.LittleEndian, appearanceTable); err != nil {
				return nil, err
			}

			if err := fp.Sync(); err != nil {
				return nil, err
			}

			if err := fp.Close(); err != nil { // Close the file so we can pin it
				return nil, err
			}

			if _, err = bl.writeBloom(chain, newTag, ToBloomPath(indexFn), false /* unused */); err != nil {
				return nil, err
			}

			// We're sucessfully written the chunk, so we don't need this any more. If the pin
			// fails we don't want to have to re-do this chunk, so remove this here.
			os.Remove(backupFn)
			return &writeChunkReport{
				Range:        base.RangeFromFilename(indexFn),
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

// Tag updates the manifest version in the chunk's header
func (chunk *Chunk) Tag(chain, newTag string, unused bool, fileName string) (err error) {
	bloomFn := ToBloomPath(fileName)
	indexFn := ToIndexPath(fileName)
	indexBackup := indexFn + ".backup"
	bloomBackup := bloomFn + ".backup"

	defer func() {
		// If the backup files still exist when the function ends, something went wrong, reset everything
		if file.FileExists(indexBackup) || file.FileExists(bloomBackup) {
			_, _ = file.Copy(bloomFn, bloomBackup)
			_, _ = file.Copy(indexFn, indexBackup)
			_ = os.Remove(bloomBackup)
			_ = os.Remove(indexBackup)
		}
	}()

	if _, err = file.Copy(indexBackup, indexFn); err != nil {
		return err
	} else if _, err = file.Copy(bloomBackup, bloomFn); err != nil {
		return err
	}

	if err = chunk.Bloom.updateTag(chain, newTag, bloomFn, unused /* unused */); err != nil {
		return err
	}

	if err = chunk.Index.updateTag(chain, newTag, indexFn, unused /* unused */); err != nil {
		return err
	}

	_ = os.Remove(indexBackup)
	_ = os.Remove(bloomBackup)

	return nil
}
