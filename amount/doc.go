// Package amount 提供链上资产金额精度换算，供 wallet-adapter 与业务下游统一使用。
//
// 严格模式（StringToBigInt / HumanToChainUnit）：
//   - 仅非负数；禁止逗号、科学计数法、正负号、空格
//   - 唯一零值 "0"；禁止 0.0、整数前导零、小数末尾零、.5、5. 等
//   - 小数位超过 decimal 直接报错，不静默截断
//   - decimal 范围 [0, MaxDecimal]（当前 MaxDecimal=78）
//
// 核心 API：
//   - StringToBigInt / BigIntToDecimal — 人类可读 ↔ 链上最小单位
//   - HumanToChainUnit / HumanToChainUnitString — 业务/合约入口
//   - SumHumanTotal — 多笔人类可读金额求和
//   - ChainUnitString — 链上最小单位十进制字符串
package amount
