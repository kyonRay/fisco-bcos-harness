---
name: fbh-setup
description: fbh harness 首次配置引导：腾讯文档授权、企微 webhook key、团队智能表格标识。新成员接入工作流、或 fbh 命令报"not configured / run /fbh-setup"错误时使用。
---

# /fbh-setup — 凭证与表格配置引导

逐项检查并引导用户补齐配置。凭证只落本机（`~/.fbh/config.json`，0600 权限），
永不入任何 git 仓库（ADR-0006）。

## 第零步：先问接入模式

- **最小接入（只治 review）**：只需 wecom_webhook + mention_map 两项，
  配完直接可用 /fbh-pr（两连）、/fbh-nudge、/fbh-review-pr、/fbh-standup。
  **跳过下面的腾讯授权和 file_id**，第四步验证改为发一条测试 nudge。
- **完整接入**：继续走全部步骤。

## 第一步：检查当前状态

```bash
fbh config show          # 三项各自 set / unset 一目了然
```

对每个 `(unset)` 项走对应小节；全部 set 则直接跳到第四步验证。

## 第二步：腾讯文档授权（token 由 tencent-docs 授权流程管理）

fbh 复用 tencent-docs skill 的授权产物（`~/.mcporter/mcporter.json`），自己不存 token。
检查方式：跑 `fbh sheet ping`，若报 "no tencent-docs token / run the tencent-docs
auth flow"，则按 tencent-docs skill 的 `references/auth.md` 流程完成授权
（`setup.sh tdoc_check_and_start_auth` → 用户浏览器授权 → `tdoc_fetch_token`）。
授权以用户本人 QQ/微信身份完成——这保证表格修改记录自带真实署名。

## 第三步：填入团队共享配置

向用户索要（通常由团队负责人在企微里发给新成员）：

```bash
fbh config set wecom_webhook '<企微群机器人 webhook 完整 URL>'
fbh config set sheet_file_id '<团队智能表格的 file_id>'
```

file_id 从表格 URL 提取：`https://docs.qq.com/smartsheet/<file_id>?...`。

团队成员 GitHub 用户名与企微 userid 不同名时（几乎必然），再配身份映射，
否则企微@落空：

```bash
fbh config set mention_map '{"<github登录名>":"<企微userid>", ...}'
```

## 第四步：验证链路

```bash
fbh sheet ping           # 应列出团队表格的全部工作表名（如 需求表 / milestone表）
```

- 成功 → 配置完成，告知用户可以开始使用 /fbh-claim、/fbh-pr 等工作流命令。
- 报 token 错误 → 回第二步。
- 报 sheet_file_id 未配置 → 回第三步。
- HTTP 401/403 → token 过期，按 tencent-docs auth.md 的过期处理重新授权。

## 注意

- 绝不把 webhook key、token、file_id 写进任何会被 commit 的文件。
- `fbh config show` 对 webhook 打码显示，可放心在共享屏幕时使用。
