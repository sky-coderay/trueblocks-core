package index

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/cache"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/file"
	"github.com/TrueBlocks/trueblocks-core/src/apps/chifra/pkg/unchained"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// HeaderRecord is the first 44 bytes of an ChunkData. This structure carries a magic number (4 bytes),
// a version specifier (32 bytes), and two four-byte integers representing the number of records in each
// of the two tables.
type HeaderRecord struct {
	Magic           uint32
	Hash            common.Hash
	AddressCount    uint32
	AppearanceCount uint32
}

func (h *HeaderRecord) String() string {
	b, _ := json.Marshal(h)
	return string(b)
}

func readHeader(fl *os.File) (header HeaderRecord, err error) {
	err = binary.Read(fl, binary.LittleEndian, &header)
	if err != nil {
		return
	}

	// Because we call this frequently, we only check that the magic number is correct
	// we let the caller check the hash if needed
	if header.Magic != file.MagicNumber {
		return header, fmt.Errorf("magic number in file %s is incorrect, expected %d, got %d", fl.Name(), file.MagicNumber, header.Magic)
	}

	return
}

func ReadChunkHeader(chain, fileName string) (header HeaderRecord, err error) {
	fileName = ToIndexPath(fileName)
	ff, err := os.Open(fileName)
	if err != nil {
		return HeaderRecord{}, err
	}
	defer ff.Close()

	if header, err = readHeader(ff); err != nil {
		return
	}

	headerHash := hexutil.Encode(header.Hash.Bytes())
	hasMagicHash := headerHash == unchained.HeaderMagicHash
	if !hasMagicHash {
		return header, fmt.Errorf("header has incorrect hash in %s, expected %s, got %s", fileName, unchained.HeaderMagicHash, headerHash)
	}

	return
}

func HasValidHeader(chain, fileName string) (bool, error) {
	header, err := ReadChunkHeader(chain, fileName)
	if err != nil {
		return false, err
	}

	rng, _ := cache.RangeFromFilename(fileName)
	if header.Magic != file.MagicNumber {
		msg := fmt.Sprintf("%s: Magic number expected (0x%x) got (0x%x)", rng, header.Magic, file.MagicNumber)
		return false, errors.New(msg)

	} else if header.Hash.Hex() != unchained.HeaderMagicHash {
		msg := fmt.Sprintf("%s: Header hash expected (%s) got (%s)", rng, header.Hash.Hex(), unchained.HeaderMagicHash)
		return false, errors.New(msg)
	}

	return true, nil
}
