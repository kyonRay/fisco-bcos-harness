---
name: fbh-pr
description: 提交 PR 全链：建 GitHub PR 指派 reviewer → 回写团队智能表格（PR链接+状态待review）→ 企微定向@ reviewer。开发完成一项子需求要提 PR 时使用。
---

# /fbh-pr — 提 PR 全链

一条命令打穿外部服务；只在当前分支已推送到远端后使用。

**三种模式**（看 `fbh config show`）：
- **最小模式**（未配表格）：两连 gh.create_pr → wecom.nudge。
- **PR 台账模式**（配了 pr_sheet_id）：建 PR 自动在台账注册一行
  （PR链接为主键，状态=待review），无需任何额外参数。
- **需求模式**（另配需求表）：再带 --table/--key 挂需求行。

**作者修复后的动作**（台账模式核心闭环）：按 review 意见改完并 push 后跑

```bash
fbh pr sync --pr <PR链接> --notify
```

状态自动从 GitHub 推导写回台账（修复中→待复审），并企微只@还没
approve 的同事继续 review。查看全队待 review 台账：`fbh pr board`。

## 流程

1. **前置确认**：分支已 push；完整模式下子需求已认领给本人（没有就先 /fbh-claim）。
2. **先 dry-run 给用户看动作序列**：

   ```bash
   # 最小模式
   fbh pr open --title "<PR 标题>" --body "<PR 描述>" \
     --reviewer <reviewer的GitHub用户名> --dry-run
   # 完整模式再加：--table <需求表sheet_id> --key "<子需求名>"
   ```

3. **用户确认后去掉 --dry-run 真跑**。
4. 把 PR URL 回显给用户。

## 结果语义

- 表格行：PR链接 已填、状态 → 待review（动作嵌入式同步，ADR-0002）。
- reviewer 在企微收到**定向@**（只@ta 本人，带 PR 链接），不是群发。
- 后续催办交给 /fbh-nudge；review 循环见 /fbh-review-pr。

## 注意

- PR 描述遵循团队惯例（FISCO-BCOS 仓库为英文描述）。
- 三步中途失败会留下部分完成状态（如 PR 已建但表格未写）：按报错单独补跑
  `fbh sheet upsert-row` / `fbh wecom nudge`，不要重跑 pr open 造出重复 PR。
