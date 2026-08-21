# ADR-0001: Harness 以本机 skills 形态运行，催办由作者侧驱动

Status: accepted
Date: 2026-08-20

## Context

团队工作流 harness 需要一个执行主体来跑催办、状态同步、门禁触发。候选：中心化常驻
bot、每个开发本机 skills、GitHub Actions、混合。团队 3-5 人，无专职运维。

同时存在一个催办动力学问题：催办的本质是在 reviewer 不作为时产生外部刺激，
但本机 skill 不自动运行，被催的人不会自己催自己。

## Decision

1. Harness 是一套分发到每个开发本机的 Claude Code skills，无中心节点。
2. 催办由**作者侧驱动**：谁的 PR 谁着急。作者本机提供 `/nudge` 命令 + 可选的
   Claude Code 定时 routine，检测自己名下 PR 无响应（如超 24h 无 review 活动）
   则自动发企微定向@。动力闭环落在利益相关者（PR 作者）身上。

## Consequences

- 无需团队服务器、无需指定运维人、无 repo secrets 管理。
- 催办的及时性依赖作者机器在线/routine 开启；作者不跑就没人催——可接受，
  因为作者正是被 review 阻塞的人。
- reviewer 侧配套 `/standup` 式命令拉取自己被指派的 review 任务，作为第二道防漏。
