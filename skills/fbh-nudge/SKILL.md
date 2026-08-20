---
name: fbh-nudge
description: 作者侧幂等催办：检测自己名下超时无 review 的 PR，企微定向@ reviewer。用户说"催一下 review"、"我的 PR 没人看"、改完想通知 reviewer 再审时使用。
---

# /fbh-nudge — 幂等催办

作者侧驱动（ADR-0001）：谁的 PR 谁着急。检测 + 发送一条命令完成：

```bash
fbh nudge run                    # 默认阈值 24h 无活动才催
fbh nudge run --threshold 4      # 自定义小时阈值
fbh nudge run --dry-run          # 只预览要发的@，不真发、不记账
```

## 判定规则（内建，无需解释给用户除非问起）

- approved 的 PR 不催。
- reviewer 已 request-changes 且作者之后无新活动 → **等的是作者**，不催 reviewer。
- 其余：最近活动距今 ≥ 阈值 → 定向@ 待 review 的 reviewer
  （无 pending reviewer 时@上次 review 的人，即"改完请复审"场景）。
- **同一 PR 同一天只发一次**（本机 state 记账；dry-run 不消耗当天额度）。

## 改完代码通知复审

作者按意见修改并 push 后，直接跑 `fbh nudge run --threshold 1` 即可——
push 刷新了活动时间，reviewer 变为催办目标。

## 可选：定时自动催办（默认不开启）

想让检测在人不在工位时也跑，可建一个 Claude Code routine / 系统级
launchd/cron 定时执行 `fbh nudge run`。幂等记账保证不会刷屏。
开启与否由每人自己决定；团队不默认开。

## 查看名下 PR 状态

```bash
fbh gh my-prs   # url / 三态 / 最近 review 时间 / reviewer 列表
```
