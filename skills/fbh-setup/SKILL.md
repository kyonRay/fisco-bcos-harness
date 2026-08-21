---
name: fbh-setup
description: fbh harness 首次配置引导：腾讯文档授权、团队智能表格标识。新成员接入工作流、或 fbh 命令报"not configured / run /fbh-setup"错误时使用。
---

# /fbh-setup — 配置引导

逐项检查并引导用户补齐配置。配置只落本机（`~/.fbh/config.json`，0600 权限），
永不入任何 git 仓库（ADR-0006）。通知走半自动播报（fbh 起草文案，作者
粘贴进企微群），不需要任何 webhook 或自动化配置。

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

PR 台账工作表 8 个**文本列**（列名须完全一致）：
`PR链接 | 标题 | 作者 | reviewers | 已approve | 状态 | 待处理人 | 更新时间`。
（openapi 无人员列类型；通知走播报文案，文本列足够。）
sheet_file_id 必须用表格 **URL 尾段**——播报里的台账链接由它拼出。

## 第四步：验证链路

```bash
fbh sheet ping           # 应列出团队表格的全部工作表名
```

- 成功 → 配置完成，告知用户可以开始使用 /fbh-pr 等工作流命令。
- 报 token 错误 → 回第二步。
- 报 sheet_file_id 未配置 → 回第三步。
- HTTP 401/403 → token 过期，按 tencent-docs auth.md 的过期处理重新授权。

## 注意

- 绝不把 token、file_id 写进任何会被 commit 的文件。
