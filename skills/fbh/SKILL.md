---
name: fbh
description: FISCO-BCOS 团队工作流 harness 总览。当用户询问 harness 的用法、命令面、或团队工作流（拆需求/认领/PR/review/门禁）如何走时使用。
---

# fbh — 团队工作流 Harness

本 harness 覆盖 需求拆分 → 认领 → 开发 → PR → AI review 循环 → 人工 review →
milestone 门禁 → 勾选完成 的全流程。腾讯智能表格是需求状态的唯一真相源，
GitHub 是 review 状态的真相源；所有外部副作用（表格/GitHub）经 `fbh` CLI
执行，`fbh <命令> --dry-run` 可预览动作而不产生副作用。

**通知全部由智能表格自动化发出**（fbh 只写表，不直发消息）：
- 新 PR 落台账行 → "添加记录提醒到群" 自动化进群通知；
- `fbh pr sync` 把"待处理人"列刷成还没 approve 的人 → "修改后提醒负责人"
  自动化定向提醒；
- 周期催办用表格的"定时提醒负责人"自动化，无需任何 fbh 命令。

**支持渐进接入**（详见 docs/usage.md）：
1. PR 台账模式（入门档）：sheet_file_id + pr_sheet_id，PR 自动注册进
   智能表格，pr sync 刷状态与待处理人，pr board 看全队欠账，不碰需求/milestone；
2. 完整模式：+需求表/milestone 表，split/claim/gate 全流程。

## 命令面

| skill | 作用 | 状态 |
|---|---|---|
| /fbh-setup | 授权、表格与自动化配置引导 | ✅ 已落地 |
| /fbh-split | 总需求拆子需求行 | ✅ 已落地 |
| /fbh-claim | 认领子需求 | ✅ 已落地 |
| /fbh-pr | 建 PR + 台账注册 + 回写需求行 | ✅ 已落地 |
| /fbh-review-pr | reviewer 侧 AI review 循环 | ✅ 已落地 |
| /fbh-standup | 防漏看板 | ✅ 已落地 |
| /fbh-gate | milestone 门禁 | ✅ 已落地 |
