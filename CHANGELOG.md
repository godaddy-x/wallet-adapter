# Changelog

## v1.0.4

- Add `types.BatchRawRequest` and `types.BatchTransferRecipient` for batch transfer build flow.
- Export batch types and `adapter.BuildBatchTransaction` in `export.go`.
- Extend `decoder.TransactionDecoder` with `CreateBatchRawTransaction` and `EstimateBatchRawTransactionFee`.

**Note:** v1.0.3 does not include batch types. Depend on `v1.0.4` or later when using `wallet-adapter-eth` batch contract decoder.
