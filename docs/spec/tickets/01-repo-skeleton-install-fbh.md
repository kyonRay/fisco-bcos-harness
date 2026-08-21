# 01 — 仓库骨架 + install 脚本 + fbh CLI 骨架

**What to build:** 团队成员 clone harness 内部仓库后运行 install 脚本，全部 skills
软链到本机 Claude Code skills 目录立即可见；`fbh` CLI 可构建、可运行，`--dry-run`
全局开关框架就位（此时尚无真实子命令业务）。升级路径 = `git pull` 即生效。
（ADR-0007）

**Blocked by:** None — can start immediately

**Status:** done (commit a7a2810 @ fisco-bcos-harness, 2026-08-20)

- [x] 新机器 clone + install 后，skills 出现在本机 skills 目录且 Claude Code 可发现
- [x] `fbh --version` 输出版本；`fbh <任意子命令> --dry-run` 框架统一拦截并打印将执行动作而不产生副作用
- [x] 工程布局遵循 fisco-bcos-release-gate 的 cmd/internal/tests 范本，tests 目录可跑通一个骨架测试
- [x] install 脚本幂等：重复运行不报错、不产生重复链接
