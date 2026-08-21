# ADR-0007: harness 放独立内部仓库，install 脚本链接到本机，升级即 git pull

Status: accepted
Date: 2026-08-20

## Context

harness 服务的是 FISCO-BCOS 这类开源仓库的开发，但自身是团队内部工具，不宜进
开源仓；放目标仓库本地目录（不入库）则无统一升级途径，5 人各自漂移。

## Decision

1. harness 全部 skills（含 review skill 的团队定制版、setup、nudge、gate、
   表格同步等）放一个**独立的团队内部 git 仓库**（可为私有 GitHub repo）。
2. 各人 clone 后运行 install 脚本，把 skills 软链接到 `~/.claude/skills/`。
3. 升级 = `git pull`（软链接使更新即时生效）。

## Consequences

- skills 版本有单一真相源，改进可走 PR review。
- 与目标开源仓库零耦合：harness 仓库地址、凭证、表格 ID 都不出现在开源仓中。
