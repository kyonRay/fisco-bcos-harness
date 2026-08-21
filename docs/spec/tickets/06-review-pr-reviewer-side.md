# 06 — /review-pr：reviewer 侧 AI review 循环与预读简报

**What to build:** reviewer 跑 `/review-pr <PR>`：复用 review skill 的 full review
流程，意见按 must-fix/should/nit 分级发 PR comment；有 must-fix 则提交 GitHub
request-changes（循环继续）；无 must-fix 则发预读简报（类型分派叙事 + 风险热图 +
AI 自报验证不了的点）作为最后一条 comment，提示 reviewer 本人聚焦 load-bearing
行，approve 由人亲手按、AI 不代按。approve 后表格行状态推进到"人工review"完成态。
含 review skill 团队定制版随 harness 分发、与个人版命名共存的处理。（ADR-0003, 0005）

**Blocked by:** 04 — /pr 全链

**Status:** done (commit 2d69636)

- [x] 有 must-fix 时：分级 comment 已发且 GitHub review 状态为 request-changes
- [x] 无 must-fix 时：预读简报 comment 含三要素（叙事/风险热图/自报盲区），且未自动 approve
- [x] 表格状态推进已实现，但时点前移：简报发出时即置"人工review"（覆盖该阶段），approve 后仅验证 review-state，"已合入"由作者合入后手动置——终审确认为有意设计
- [x] 团队定制版 review skill 与个人已装版本可共存，无命名冲突
