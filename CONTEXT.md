# CONTEXT — 团队工作流 Harness

本文件是工作流 harness 设计的术语表。写 issue 标题、spec、skill 文档时用这里的
词，不要漂移到同义词。

## 术语

- **Harness（工作流骨架）**: 分发到每个开发本机的一套 Claude Code skills，
  覆盖 需求拆分 → 开发 → PR → AI review 循环 → 人工 review → 门禁 → 勾选完成
  的全流程。无中心节点（ADR-0001）。
- **真相源（source of truth）**: 腾讯智能表格。需求拆分、认领、进度、完成勾选
  的唯一权威状态；GitHub 是镜像（ADR-0002）。
- **动作嵌入式同步**: 状态变更者的 skill 在动作尾部顺带回写表格，无轮询无
  webhook（ADR-0002）。
- **催办（nudge）**: 作者侧驱动的提醒动作——`/nudge` 命令或定时 routine 检测
  自己名下 PR 无响应后发企微定向@（ADR-0001）。区别于"群发通知"。
- **AI review 循环**: reviewer 的 AI agent 提意见 → 作者修改 → 再审的循环；
  终止条件 = reviewer 侧 AI 无 must-fix（ADR-0003）。
- **must-fix / should / nit**: AI review 意见的三级分级。只有 must-fix 阻塞循环
  终止。
- **人工聚焦提示 / 预读简报**: AI review 循环终止后，reviewer AI 以 review skill
  预读简报形态发出的最后一条 PR comment：类型分派叙事 + 风险热图 +
  AI 自报验证不了的点（ADR-0005）。
- **风险热图**: 把 diff 路由为 mechanical（skim 即可）与 load-bearing（逐行看）
  两类，load-bearing 条目按 `path:line` 排序、每条一句为什么险（源自 review
  skill Stage 0）。
- **门禁（gate）**: milestone 达成时由该 milestone 负责人本机跑 `/gate` 的完整
  集成测试 + SIT，形态参考 fisco-bcos-release-gate skill；通过则回写勾选，
  失败则自动建缺陷子需求行并企微@相关 PR 作者（ADR-0004）。
- **认领人**: 智能表格中子需求行的负责人字段，也是 PR 作者。
- **milestone 负责人**: 智能表格中 milestone 行的负责人字段，门禁的触发者与
  失败分诊兜底人（ADR-0004）。
- **harness 仓库**: 存放全部 harness skills 的独立团队内部 git 仓库；install
  脚本软链到 `~/.claude/skills/`，升级即 git pull（ADR-0007）。凭证不入库，
  各自本机按模板配置（ADR-0006）。
