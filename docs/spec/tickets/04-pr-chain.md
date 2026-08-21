# 04 — /pr 全链：建 PR → 回写表格 → 企微定向@

**What to build:** PR 作者跑 `/pr`：创建 GitHub PR 并指派 reviewer，skill 尾部
自动把 PR 链接回写到表格对应子需求行、状态推进到"待review"，并向被指派 reviewer
发企微定向@（带 PR 链接，非群发）。这是第一条打穿 表格+GitHub+企微 三个外部
服务的完整链路。（ADR-0001, 0002）

**Blocked by:** 03 — 表格 schema + /split 拆分 + /claim 认领

**Status:** done (commit 22a64ba; 真实 gh/企微终验待用户)

- [x] `/pr` 一次执行后：GitHub 上 PR 存在且 reviewer 已指派；表格行 PR 链接与状态已更新；企微群收到恰一条定向@（mention 列表只含 reviewer）
- [x] `fbh wecom nudge` 落地，fake endpoint 测试断言 mention 列表与消息含 PR 链接
- [x] `fbh gh review-state` 落地，gh JSON 固件回放测试覆盖 approve / request-changes / 无 review 三态判定
- [x] 全链支持 dry-run：打印三个外部动作的完整序列而零副作用
