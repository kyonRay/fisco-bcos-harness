# ADR-0004: milestone 门禁由负责人手动触发，失败自动建缺陷子需求行

Status: accepted
Date: 2026-08-20

## Context

milestone 门禁（完整集成测试 + SIT，形态参考 fisco-bcos-release-gate skill）跑一轮
耗时长、占满一台机器。本机 skills 形态（ADR-0001）下没有 CI 机器池。

## Decision

1. 智能表格中每个 milestone 行有**负责人**字段；到达 milestone 后由负责人在自己
   机器上跑 `/gate` 命令。
2. 门禁通过：`/gate` 尾部回写表格，勾选该 milestone 及其下子需求的完成状态
   （动作嵌入式同步，ADR-0002）。
3. 门禁失败：`/gate` 自动在表格新建缺陷子需求行（挂在该 milestone 下），并发
   企微定向@相关 PR 作者；milestone 状态保持未完成。

## Consequences

- 无需常驻测试机；代价是门禁期间负责人机器被占用。
- "相关 PR 作者"的归因由门禁工具尽力而为（失败场景 → 涉及模块 → 该 milestone
  内改过该模块的 PR）；归因不出来就@负责人自己分诊。
