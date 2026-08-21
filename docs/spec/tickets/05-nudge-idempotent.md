# 05 — /nudge 幂等催办 + 可选定时 routine

**What to build:** PR 作者跑 `/nudge`：检测自己名下所有等 review 的 PR，对超过
阈值（默认 24h，可配置）无 review 活动的发企微定向@被指派 reviewer；同一 PR
同一天不重复@（幂等）。作者按意见修改推送后同一命令用于通知再审。提供可选的
Claude Code 定时 routine 配置方式，让检测在人不在工位时也运行。催办动力闭环在
作者侧——谁的 PR 谁着急。（ADR-0001）

**Blocked by:** 04 — /pr 全链

**Status:** done (commit 3694426)

- [x] dry-run 测试：PR 挂 25h 无响应 → 输出恰一条企微请求；同日第二次跑 → 零输出
- [x] `fbh gh my-prs` 落地：列出本人名下开放 PR 及各自最近 review 活动时间（固件回放测试）
- [x] reviewer 已 request-changes 且作者未推新 commit 的 PR 不催（等的是作者不是 reviewer）
- [x] 阈值可配置；routine 的开启步骤有文档且默认不开启
