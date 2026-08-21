# ADR-0003: AI review 循环的终止判定以 reviewer 侧为准，落 GitHub 原生 review 状态

Status: accepted
Date: 2026-08-20

## Context

AI review 循环（AI 提意见 → 作者改 → 再审）的终止条件必须机器可判定，否则催办
自动化无法知道何时进入"人工 review 阶段"。两个开发的 AI agent 可能结论矛盾。
结论如果埋在 AI 会话里，看板/表格同步都读不到。

## Decision

1. **reviewer 侧 AI 说了算**：只有被指派 reviewer 的 AI agent 的结论计入终止判定；
   作者自己的 AI 自审只是提 PR 前的前置门槛。
2. **结论载体是 GitHub 原生 review 状态**：reviewer AI 审完无 must-fix 时，由
   reviewer 确认后提交 GitHub approve；有 must-fix 则 request-changes。
   `gh api` 直接可查，零自定义状态解析。

## Consequences

- 催办/看板/表格同步只需查 GitHub review state，不解析 comment 文本。
- AI 的具体意见仍以 PR comment 呈现（含 must-fix/should/nit 分级），但机器判定
  只看 review state，意见内容给人看。
- "AI 无 must-fix → 提示人工 review 聚焦点"发生在 approve 之前：approve 是
  人工 review 完成后由 reviewer 本人做出的动作，AI 不代按。
