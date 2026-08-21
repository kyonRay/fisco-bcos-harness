# ADR-0005: 人工聚焦提示复用 review skill 的预读简报形态（风险热图）

Status: accepted
Date: 2026-08-20

## Context

AI review 循环终止（reviewer 侧 AI 无 must-fix，ADR-0003）后，需要提示 reviewer
本人把精力投到哪些代码。若提示只是 AI 复述自己已审内容，人工 review 就成走过场。

本地 `review` skill（~/.cc-switch/skills/review/SKILL.md）的 Stage 0 已有成熟形态：
**预读简报** = 类型分派叙事（fix/feature/refactor 各自的问题→根因→修法 /
文件角色→入口→逐函数图）+ **风险热图** + CI 分诊。风险热图把 diff 路由为
mechanical（签名穿透、rename、fixture 适配——"skim 即可，我查过一致性"）与
load-bearing（并发、生命周期、共识、存储、密码学、序列化——逐行看），
按 `path:line` 排序、每条一句为什么险。

## Decision

人工聚焦提示 = review skill 的预读简报，harness 不另造格式：

1. reviewer 侧 AI 在 review 循环中跑 full review（Stage 1-4），无 must-fix 后
   产出预读简报作为最后一条 PR comment。
2. 简报必须包含：类型分派叙事、风险热图（load-bearing 条目排序 + 单句理由）、
   以及 **AI 明确说自己验证不了的点**（需要业务/领域知识的判断、未能实证的
   并发时序假设、测试未覆盖的路径）——这是人工不可替代的部分，不许省略。
3. reviewer 本人读完简报、看完 load-bearing 行后自己提交 GitHub approve
   （ADR-0003：AI 不代按）。

## Consequences

- 团队每人本机需安装 review skill（随 harness 仓库一起分发，ADR-0007）。
- 简报作为 PR comment 留档，事后可审计"人工 review 当时被提示看了什么"。
