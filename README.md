# fisco-bcos-harness

FISCO-BCOS 团队工作流 harness：一套分发到每个开发本机的 Claude Code skills +
`fbh` CLI，覆盖 需求拆分 → 认领 → PR → AI review 循环 → 人工 review →
milestone 门禁 的全流程。设计决策见团队 spec 与 ADR-0001~0007。

## 安装 / 升级

```bash
git clone <this-repo> && cd fisco-bcos-harness
./install.sh        # 软链 skills 到 ~/.claude/skills/ 并构建 bin/fbh
git pull            # 升级；软链使 skills 即时生效，重跑 install.sh 重建 fbh
```

## fbh CLI

skills 触碰外部服务（腾讯智能表格 / 企微 webhook / GitHub）的唯一通道。
任何子命令加 `--dry-run` 输出将执行的动作 JSON 行而不产生副作用：

```bash
bin/fbh --version
bin/fbh <command> --dry-run
```

## 使用文档

团队使用文档（安装/角色流程/命令参考/状态机/排查）见 [docs/usage.md](docs/usage.md)。

## 测试

```bash
go test ./...            # CLI 行为测试（dry-run 动作流等）
bash tests/install_test.sh   # install 幂等性
```

## 布局

- `cmd/fbh/` — CLI 入口薄壳
- `internal/cli/` — 命令注册表、dry-run 路由
- `internal/action/` — 外部副作用的动作描述
- `skills/` — 分发给团队的 Claude Code skills（install.sh 逐个软链）
- `tests/` — 跨语言行为测试脚本
