// Package amount on-chain asset amount precision conversion for wallet-adapter and downstream business use.
//
// Strict mode (StringToBigInt / HumanToChainUnit):
//   - non-negative only; commas, scientific notation, signs, and spaces are rejected
//   - sole zero value "0"; rejects 0.0, integer leading zeros, trailing fractional zeros, .5, 5., etc.
//   - fractional digits exceeding decimal return error; no silent truncation
//   - decimal range [0, MaxDecimal] (currently MaxDecimal=78)
//
// Core API:
//   - StringToBigInt / BigIntToDecimal — human-readable ↔ smallest on-chain unit
//   - HumanToChainUnit / HumanToChainUnitString — business/contract entry points
//   - SumHumanTotal — sum multiple human-readable amounts
//   - ChainUnitString — decimal string of smallest on-chain unit
package amount
