---
name: fbh-pr
description: 提交 PR 全链：建 GitHub PR 指派 reviewer → 自动注册进 PR 台账 → 输出播报文案供作者粘贴到企微群 → 可选回写需求行。开发完成一项子需求要提 PR 时使用。
---

# /fbh-pr — 提 PR 全链

一条命令打穿外部服务；只在当前分支已推送到远端后使用。通知走半自动
播报：命令末尾输出"---- 复制到企微群 ----"文案块，**把它原样递给用户，
提醒 ta 粘贴进群**——fbh 起草，人来送达。

**模式**（看 `fbh config show`）：
- **PR 台账模式**（配了 pr_sheet_id）：建 PR 自动在台账注册一行
  （PR链接为主键，状态=待review，待处理人=全部 reviewer），无需额外参数。
- **需求模式**（另配需求表）：再带 --table/--key 挂需求行。

**作者修复后的动作**（台账模式核心闭环）：按 review 意见改完并 push 后跑

```bash
fbh pr sync --pr <PR链接>
```

状态自动从 GitHub 推导写回台账（修复中→待复审），**待处理人列刷成还没
approve 的人**，播报文案只点名他们——把文案递给用户粘群即完成催复审。
合入后再跑一次同样命令，状态变已合入、待处理人清空（不再出播报）。
查看全队待 review 台账：`fbh pr board`。

## 流程

1. **前置确认**：分支已 push；完整模式下子需求已认领给本人（没有就先 /fbh-claim）。
2. **先 dry-run 给用户看动作序列**：

   ```bash
   fbh pr open --title "<PR 标题>" --body "<PR 描述>" \
     --reviewer <reviewer的GitHub用户名> --dry-run
   # 完整模式再加：--table <需求表sheet_id> --key "<子需求名>"
   ```

3. **用户确认后去掉 --dry-run 真跑**（dry-run 只出动作行，不出播报）。
4. 把 PR URL 和**播报文案块**一起回显给用户，提醒 ta 粘到企微群。

## 结果语义

- 台账行：PR链接/标题/作者/reviewers/状态=待review/待处理人 已填。
- 群通知靠用户粘贴播报文案；review 循环见 /fbh-review-pr。

## 注意

- PR 描述遵循团队惯例（FISCO-BCOS 仓库为英文描述）。
- 中途失败会留下部分完成状态（如 PR 已建但台账未写）：直接跑
  `fbh pr sync --pr <PR链接>` 补齐台账行，不要重跑 pr open 造出重复 PR。
