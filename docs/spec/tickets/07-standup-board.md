# 07 — /standup 防漏看板

**What to build:** 团队任何成员早上跑 `/standup`，一屏看到：自己名下等 review 的
PR（含挂了多久、催过几次）+ 自己被指派但尚未完成的 review 任务清单（含 PR 链接
与当前循环状态）。这是催办之外的第二道防漏线——漏了能一眼看见。（ADR-0001）

**Blocked by:** 05 — /nudge 幂等催办

**Status:** done (commit 8392ef3)

- [x] `/standup` 输出两节：我的待 review PR / 我欠的 review，各含可点开的 PR 链接
- [x] `fbh gh my-reviews` 落地：列出指派给本人且未 approve 的 PR（固件回放测试）
- [x] 循环状态正确区分"等 reviewer 首审 / 等作者改 / 等 reviewer 复审 / 等人工 approve"
- [x] 两节皆空时输出明确的"无欠账"而非空白
