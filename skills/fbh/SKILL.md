---
name: fbh
description: FISCO-BCOS 团队工作流 harness 总览。当用户询问 harness 的用法、命令面、或团队工作流（拆需求/认领/PR/催办/review/门禁）如何走时使用。
---

# fbh — 团队工作流 Harness

本 harness 覆盖 需求拆分 → 认领 → 开发 → PR → AI review 循环 → 人工 review →
milestone 门禁 → 勾选完成 的全流程。腾讯智能表格是需求状态的唯一真相源，
GitHub 是镜像；所有外部副作用（表格/企微/GitHub）经 `fbh` CLI 执行，
`fbh <命令> --dry-run` 可预览动作而不产生副作用。

## 命令面

| skill | 作用 | 状态 |
|---|---|---|
| /fbh-setup | 凭证与表格配置引导 | ✅ 已落地 |
| /fbh-split | 总需求拆子需求行 | ✅ 已落地 |
| /fbh-claim | 认领子需求 | ✅ 已落地 |
| /fbh-pr | 建 PR + 回写表格 + 企微@reviewer | ✅ 已落地 |
| /fbh-nudge | 幂等催办 | ✅ 已落地 |
| /fbh-review-pr | reviewer 侧 AI review 循环 | ✅ 已落地 |
| /fbh-standup | 防漏看板 | ✅ 已落地 |
| /fbh-gate | milestone 门禁 | ✅ 已落地 |
