# Changelog

## 生产批量路径说明（2026-06-27）

- **生产批量转账** 使用 open_scanner `CreateBatchTransferTrade` → `SmartContractRawTransaction`（CLI type=2），**不再**通过 `adapter.BuildBatchTransaction`（v1.0.4 历史 API）。
- 遗留 type=0 批量 pending 仍可由 wallet-adapter-eth 做交叉校验。

## v1.0.5

- Add public package `amount` for human-readable ↔ on-chain smallest-unit conversion (`StringToBigInt`, `BigIntToDecimal`, `HumanToChainUnit`, `SumHumanTotal`).
- **Constraints**: All functions only accept non-negative amounts (>= 0). Negative inputs return errors.
- **API Change**: `BigIntToDecimal` now returns `(string, error)` instead of `string` to handle negative and nil inputs.
- **Tests**: Comprehensive test coverage including boundary cases (zero, nil, decimals=0, scientific notation, multiple dots, empty string).
- Downstream services should import `github.com/godaddy-x/wallet-adapter/amount` instead of copying conversion logic.

## v1.0.4

- Add `types.BatchRawRequest` and `types.BatchTransferRecipient` for batch transfer build flow.
- Export batch types and `adapter.BuildBatchTransaction` in `export.go`.
- Extend `decoder.TransactionDecoder` with `CreateBatchRawTransaction` and `EstimateBatchRawTransactionFee`.

**Note:** v1.0.3 does not include batch types. Depend on `v1.0.4` or later when using `wallet-adapter-eth` batch contract decoder.
