package async

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	aetraaddress "github.com/sovereign-l1/l1/app/addressing"
)

func DeriveContractAddress(deployer sdk.AccAddress, codeHash []byte, salt []byte) (sdk.AccAddress, error) {
	if err := aetraaddress.RejectZeroAddress("contract deployer", deployer); err != nil {
		return nil, err
	}
	if len(codeHash) != CodeHashLength {
		return nil, fmt.Errorf("contract code hash must be %d bytes", CodeHashLength)
	}
	h := sha256.New()
	writePart(h.Write, []byte(AddressDerivationDomain))
	writePart(h.Write, deployer)
	writePart(h.Write, codeHash)
	writePart(h.Write, salt)
	return sdk.AccAddress(h.Sum(nil)), nil
}

func writePart(write func([]byte) (int, error), bz []byte) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len(bz)))
	_, _ = write(length[:])
	_, _ = write(bz)
}
