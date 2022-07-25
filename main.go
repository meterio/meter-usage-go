package main

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/meterio/meter-pov/genesis"
	"github.com/meterio/meter-pov/meter"
	"github.com/meterio/meter-pov/script"
	"github.com/meterio/meter-pov/script/auction"
	"github.com/meterio/meter-pov/tx"
)

func buildAuctionBuildTx() (*tx.Transaction, error) {
	chainTag := byte(82)
	bestNum := uint32(26360919) // TODO: use latest block number
	builder := new(tx.Builder)
	builder.ChainTag(chainTag).
		BlockRef(tx.NewBlockRef(bestNum)).
		Expiration(32).
		GasPriceCoef(0).
		Gas(meter.BaseTxGas * 2). // buffer for builder.Build().IntrinsicGas()
		DependsOn(nil).
		Nonce(5) // TODO: please use random int

	// prepare auction bid data
	body := &auction.AuctionBody{
		Opcode:      auction.OP_BID,
		Version:     uint32(0),
		StartHeight: 0,
		StartEpoch:  0,
		EndHeight:   0,
		EndEpoch:    0,
		Sequence:    0,
		AuctionID:   meter.Bytes32{},
		Timestamp:   uint64(time.Now().Unix()),
		Nonce:       rand.Uint64(),
	}
	payload, err := rlp.EncodeToBytes(body)
	if err != nil {
		fmt.Println("payload rlp error:", err.Error())
		return nil, err
	}

	s := &script.Script{
		Header: script.ScriptHeader{
			Version: uint32(0),
			ModID:   script.AUCTION_MODULE_ID,
		},
		Payload: payload,
	}
	data, err := rlp.EncodeToBytes(s)
	if err != nil {
		fmt.Println("script data rlp error:", err.Error())
		return nil, err
	}
	data = append(script.ScriptPattern[:], data...)
	prefix := []byte{0xff, 0xff, 0xff, 0xff}
	data = append(prefix, data...)

	// build tx
	builder.Clause(
		tx.NewClause(&auction.AuctionAccountAddr).
			WithValue(big.NewInt(0)).
			WithToken(meter.MTRG).
			WithData(data))
	tx := builder.Build()

	// sign the tx
	sig, err := crypto.Sign(tx.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	if err != nil {
		fmt.Println("sign error: ", err)
		return nil, err
	}
	tx = tx.WithSignature(sig)
	return tx, nil
}

func main() {
	fmt.Println(auction.AuctionAccountAddr)
	tx, err := buildAuctionBuildTx()
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Build Tx: ", tx)
	}
}
