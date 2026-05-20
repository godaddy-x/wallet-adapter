// Package amount 提供链上资产金额精度换算，供 wallet-adapter-eth 与业务下游统一使用。
//
// 核心 API：
//   - StringToBigInt / BigIntToDecimal — 人类可读 ↔ 链上最小单位（与历史 adapter 行为一致）
//   - HumanToChainUnit / HumanToChainUnitString — 严格正数换算（拒绝超精度非零小数）
//   - AssertRoundtrip / SumHumanTotal — 批量场景 value 与 amounts 总和一致性校验
package amount
