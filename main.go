package main

import (
	"context"
	"fmt"
	"math/big"
	"math/rand"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/meterio/meter-pov/meter"
	"github.com/meterio/meter-pov/script"
	"github.com/meterio/meter-pov/script/auction"
)

func buildAuctionBidTx(privKeyHex string, amount *big.Int, nonce uint64) (*types.Transaction, error) {
	privKey, _ := crypto.HexToECDSA(privKeyHex)
	owner := crypto.PubkeyToAddress(privKey.PublicKey)

	chainID := big.NewInt(82)

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
		Bidder:      meter.MustParseAddress(owner.String()),
		Timestamp:   uint64(time.Now().Unix()),
		Amount:      amount,
		Nonce:       rand.Uint64(),
	}
	payload, err := rlp.EncodeToBytes(body)
	if err != nil {
		fmt.Println("payload rlp error:", err.Error())
		return nil, err
	}

	s := &script.Script{
		Header:  script.ScriptHeader{Version: uint32(0), ModID: script.AUCTION_MODULE_ID},
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

	gasLimit := uint64(250000)
	gasPrice := big.NewInt(50000000000)
	tx := types.NewTransaction(nonce, owner, amount, gasLimit, gasPrice, data)

	// sign the tx
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	if err != nil {
		fmt.Println(err)
	}

	return signedTx, nil
}

func main() {
	fmt.Println(auction.AuctionAccountAddr)
	privKeyHex := "eac5...9a2f"                                   // TODO: replace with actual private key
	amount := big.NewInt(0).Mul(big.NewInt(13), big.NewInt(1e18)) // TODO: replace with actual amount
	nonce := uint64(3)                                            // TODO: use random number
	tx, err := buildAuctionBidTx(privKeyHex, amount, nonce)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	client, err := ethclient.Dial("http://rpc.meter.io:8545")
	if err != nil {
		fmt.Println("Could not connect to rpc endpoint:", err)
		return
	}
	err = client.SendTransaction(context.Background(), tx)
	if err != nil {
		fmt.Println("Could not send transaction: ", err)
		return
	}
	fmt.Println("Sent tx: ", tx.Hash())
}
