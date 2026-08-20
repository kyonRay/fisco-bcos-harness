---
name: fbh-setup
description: fbh harness 首次配置引导：腾讯文档授权、团队智能表格标识、表格自动化提醒规则。新成员接入工作流、或 fbh 命令报"not configured / run /fbh-setup"错误时使用。
---

# /fbh-setup — 配置引导

逐项检查并引导用户补齐配置。配置只落本机（`~/.fbh/config.json`，0600 权限），
永不入任何 git 仓库（ADR-0006）。通知不经 fbh —— 由智能表格的自动化规则
发出，所以本引导的最后一步是核对表格侧的自动化配置。

## 第零步：先问接入档

- **PR 台账接入（入门档，只治 review）**：腾讯授权 + sheet_file_id +
  pr_sheet_id。提 PR 自动注册台账、`fbh pr sync` 刷状态与待处理人、
  `fbh pr board` 看全队欠账。
- **完整接入**：台账档 + 需求表/milestone 表（/fbh-split、/fbh-claim、
  /fbh-gate 用）；milestone 负责人另配 gate_cmd。

## 第一步：检查当前状态

```bash
fbh config show          # 各项 set / unset 一目了然
```

对每个 `(unset)` 项走对应小节；全部 set 则直接跳到验证。

## 第二步：腾讯文档授权（token 由 tencent-docs 授权流程管理）

fbh 复用 tencent-docs skill 的授权产物（`~/.mcporter/mcporter.json`），自己不存 token。
检查方式：跑 `fbh sheet ping`，若报 "no tencent-docs token / run the tencent-docs
auth flow"，则按 tencent-docs skill 的 `references/auth.md` 流程完成授权
（`setup.sh tdoc_check_and_start_auth` → 用户浏览器授权 → `tdoc_fetch_token`）。
授权以用户本人 QQ/微信身份完成——这保证表格修改记录自带真实署名。

## 第三步：填入团队共享配置

```bash
fbh config set sheet_file_id '<团队智能表格的 file_id>'
fbh config set pr_sheet_id  '<PR台账工作表的 sheet_id>'   # fbh sheet ping 可查
```

file_id 从表格 URL 提取：`https://docs.qq.com/smartsheet/<file_id>?...`。

PR 台账工作表 8 列（列名须完全一致）：
`PR链接 | 标题 | 作者 | reviewers | 已approve | 状态 | 待处理人 | 更新时间`。
其中 **待处理人 建为人员列**（自动化按它定向提醒），其余为文本列。

## 第四步：核对表格自动化（团队一次性，通常由负责人配）

在表格右上角"自动化"面板确认这三条规则存在（fbh 不发通知，全靠它们）：

1. **添加记录提醒到群**：台账新增记录 → 发消息到团队群（新 PR 通知）。
2. **修改后提醒负责人**：台账记录变更 → 发文档消息给 [待处理人]
   （修复后催未 approve 的人；能加条件则限定 状态=待复审，省触发次数）。
3. **定时提醒负责人**（可选）：每天/每周定时给 [待处理人] 发文档消息（周期催办）。

注意自动化有每月运行次数配额（免费 100 次/月），不够用时找管理员提额。

## 第五步：验证链路

```bash
fbh sheet ping           # 应列出团队表格的全部工作表名
```

- 成功 → 配置完成，告知用户可以开始使用 /fbh-pr 等工作流命令。
- 报 token 错误 → 回第二步。
- 报 sheet_file_id 未配置 → 回第三步。
- HTTP 401/403 → token 过期，按 tencent-docs auth.md 的过期处理重新授权。

## 注意

- 绝不把 token、file_id 写进任何会被 commit 的文件。
- 待处理人列若建成了文本列，自动化的"提醒负责人"选不到它——务必人员列。
