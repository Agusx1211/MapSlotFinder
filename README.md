
# Solidity Map Slot finder

This is a simple tool to find the slot ID of a given map in Solidity, without having access to the source code, it uses brute-force to find the slot ID.
The tool assumes that the storage layout is the same as the one used by the Solidity compiler.

See more: [https://docs.soliditylang.org/en/latest/internals/layout_in_storage.html#mappings-and-dynamic-arrays](https://docs.soliditylang.org/en/latest/internals/layout_in_storage.html#mappings-and-dynamic-arrays)

## Usage

```bash
go run ./src/main.go [provider] [address] [reference] [key0] [key1] [key2] ...
```

### Example

We will use Arbitrum Nova USDC contract as an example. But this tool can be used for any contract.

[0x750ba8b76187092b0d1e87e28daaf484d1b5273b](https://nova.arbiscan.io/token/0x750ba8b76187092b0d1e87e28daaf484d1b5273b)

If we want to find the map slot for `balanceOf` we need to need to take an address that has balance, we pass the balance as reference and the address as the key.

In this case, we will find the `balanceOf` slot because the `0x7af288415d718761e9ba2ba74c5c838437ae0ae5` address has a balance of `4017357`.

```bash
go run src/main.go https://nodes.sequence.app/arbitrum-nova 0x750ba8b76187092b0d1e87e28daaf484d1b5273b 4017357 0x7af288415d718761e9ba2ba74c5c838437ae0ae5
> Slot found: 51
```

If we use a different address, we will get the same slot ID.

```bash
go run src/main.go https://nodes.sequence.app/arbitrum-nova 0x750ba8b76187092b0d1e87e28daaf484d1b5273b 0x0c 0x34430eb654ae2a20fa7c281548ef3ce665d58db0
> Slot found: 51
```

We can also use it to find other slots, for example the `nonces` map. To find a different map we need to pass a different reference.

The nonce for the `0x7af288415d718761e9ba2ba74c5c838437ae0ae5` address is `38`.

```bash
go run src/main.go https://nodes.sequence.app/arbitrum-nova 0x750ba8b76187092b0d1e87e28daaf484d1b5273b 38 0x7af288415d718761e9ba2ba74c5c838437ae0ae5
> Slot found: 153
```

> Notice that if the reference is not unique, the tool will return a list of possible slots. In this case, tweak the reference to make it unique.
