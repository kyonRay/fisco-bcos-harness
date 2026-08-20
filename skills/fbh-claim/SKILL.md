---
name: fbh-claim
description: 认领团队智能表格中的一条子需求（置认领人、状态推进到开发中）。开发开始一项新工作、说"我来做 X"时使用。
---

# /fbh-claim — 认领子需求

## 流程

1. 确认要认领的需求名（用户没说清就先 `fbh sheet ping` 找到需求表、
   列出待认领行让用户挑）。
2. 认领（owner 用团队约定的名字，通常是 GitHub 用户名）：

   ```bash
   fbh sheet claim --table <需求表sheet_id> --key "<子需求名>" --owner "<我>"
   ```

3. 结果处理：
   - 成功 → 表格行已置 认领人 + 状态=开发中，告知用户可以开分支开工。
   - 报 "already claimed by X" → 该需求已被 X 认领，提示用户找 X 协调，
     **不要**用 upsert-row 强抢。
   - 报 "not found" → 需求名打错或还没拆分，引导用户先跑 /fbh-split。

## 注意

认领即承诺：后续 /fbh-pr、/fbh-nudge 都以认领人=PR 作者为前提。
