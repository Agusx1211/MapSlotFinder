package main

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/0xsequence/ethkit/ethcoder"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

type CallOverride struct {
	Code string
}

// Source: StorageFetcher.huff
const FetcherProgram = "0x60005b803554815260200136811061000257366000f3"

func FetchSlots(ctx context.Context, provider *ethrpc.Provider, address common.Address, slots [][32]byte) ([][32]byte, error) {
	// Generate the call data
	// encoding all slots one after the other
	calldata := make([]byte, 0, len(slots)*32)
	for _, slot := range slots {
		calldata = append(calldata, slot[:]...)
	}

	// Call the address, but set the code override to the fetcher program
	type Call struct {
		To   common.Address `json:"to"`
		Data string         `json:"data"`
	}

	estimateCall := &Call{
		To:   address,
		Data: "0x" + common.Bytes2Hex(calldata),
	}

	var res string
	rpcCall := ethrpc.NewCallBuilder[string]("eth_call", nil, estimateCall, nil, map[common.Address]*CallOverride{
		address: {Code: FetcherProgram},
	})
	_, err := provider.Do(ctx, rpcCall.Into(&res))
	if err != nil {
		return [][32]byte{}, err
	}

	resBytes := common.FromHex(res)
	if len(resBytes) != len(slots)*32 {
		return [][32]byte{}, fmt.Errorf("fetcher: unexpected response length")
	}

	// Decode the response
	results := make([][32]byte, len(slots))
	for i := 0; i < len(slots); i++ {
		copy(results[i][:], resBytes[i*32:(i+1)*32])
	}

	return results, nil
}

func showHelp() {
	fmt.Println("use: map-slot-finder [provider] [address] [reference] [key0] [key1] [key2] ...")
}

func intOrHexToBytes32(val string) [32]byte {
	// If starts with 0x then it's a hex string
	// if not it is a decimal string
	if val[:2] == "0x" {
		full := common.FromHex(val)
		if len(full) > 32 {
			panic(fmt.Errorf("value too large: %s", val))
		}

		var result [32]byte
		copy(result[:], common.LeftPadBytes(full, 32))
		return result
	}

	// Parse as integer
	bi, ok := big.NewInt(0).SetString(val, 10)
	if !ok {
		panic(fmt.Errorf("invalid integer: %s", val))
	}

	full := bi.Bytes()
	if len(full) > 32 {
		panic(fmt.Errorf("value too large: %s", val))
	}

	var result [32]byte
	copy(result[:], common.LeftPadBytes(full, 32))
	return result
}

func main() {
	args := os.Args[1:]

	if len(args) < 4 {
		showHelp()
		return
	}

	providerUrl := args[0]
	provider, err := ethrpc.NewProvider(providerUrl)
	if err != nil {
		panic(err)
	}

	if !common.IsHexAddress(args[1]) {
		panic(fmt.Errorf("invalid address: %s", args[1]))
	}

	address := common.HexToAddress(args[1])
	reference := intOrHexToBytes32(args[2])

	// Compute the "prefix"
	// (all the keys one after the other)
	prefix := make([]byte, 0, len(args[3:])*32)
	for _, key := range args[3:] {
		p := intOrHexToBytes32(key)
		prefix = append(prefix, p[:]...)
	}

	slots := make([][32]byte, 32768)
	for i := 0; i < 32768; i++ {
		abiEncoded := make([]byte, 0, len(prefix)+32)
		abiEncoded = append(abiEncoded, prefix...)
		abiEncoded = append(abiEncoded, common.LeftPadBytes(common.FromHex(fmt.Sprintf("%x", i)), 32)...)
		h := ethcoder.Keccak256(abiEncoded)
		copy(slots[i][:], h[:])
	}

	// Fetch the slots
	ctx := context.Background()
	results, err := FetchSlots(ctx, provider, address, slots)
	if err != nil {
		panic(err)
	}

	// Find all results
	matches := make([]int, 0)
	for i, result := range results {
		if result == reference {
			matches = append(matches, i)
		}
	}

	if len(matches) == 0 {
		fmt.Println("no matches found")
		return
	}

	if len(matches) == 1 {
		fmt.Printf("Slot found: %d\n", matches[0])
		return
	}

	fmt.Printf("Multiple matches found: %v\n", matches)
}
