# Spec: 团队工作流 Harness

Status: ready-for-agent
Date: 2026-08-20
依据: docs/adr/0001–0007, CONTEXT.md 术语表（术语用法以 CONTEXT.md 为准）

## Problem Statement

一个 3-5 人团队在 FISCO-BCOS 这类开源仓库上协作开发。总需求拆成子需求分工后，
流程里有三处持续失血：

1. **催 review 靠人肉**：PR 推上去后在企微群喊一嗓子，容易被漏看，作者只能反复催。
2. **AI review 循环没有可判定的终点**：各开发的 AI agent 提意见→修改→再催，
   什么时候算"AI 这关过了、该人上了"没有机器可查的状态，人工 review 也不知道
   该把精力聚焦到哪些代码，容易走过场。
3. **需求状态散落**：拆分/认领/进度在一处，PR 状态在 GitHub，milestone 测试结果
   又在别处，勾选"完成"靠人对着几个窗口核对。

## Solution

一套分发到每个开发本机的 Claude Code skills（harness），存放在独立内部仓库、
软链安装、`git pull` 升级。腾讯智能表格是需求状态的唯一真相源，GitHub 是镜像，
每个改状态的 skill 在动作尾部顺带回写表格（动作嵌入式同步）。催办由作者侧驱动
（`/nudge` + 可选定时 routine，企微定向@）。AI review 循环的终止判定 =
被指派 reviewer 的 AI 无 must-fix，结论落 GitHub 原生 review 状态；终止后 AI 发
预读简报（叙事 + 风险热图 + 自报盲区）指引人工 review，approve 由人亲手按。
milestone 门禁由负责人本机跑 `/gate`（复用 fisco-bcos-release-gate 形态），
通过回写勾选，失败自动建缺陷子需求行并企微@相关 PR 作者。

所有外部副作用（腾讯表格 API、企微 webhook、GitHub 查询）收进单一 helper CLI
（工作名 `fbh`），skills 只调它——这是整个 harness 唯一的测试接缝。

## User Stories

1. 作为需求负责人，我想把一项总需求拆成多个子需求行写进智能表格（含认领人、
   关联 milestone 字段），以便团队分工有单一权威视图。
2. 作为开发（认领人），我想在表格认领子需求后开工，认领动作自动记录到表格，
   以便别人不会重复认领。
3. 作为 PR 作者，我想在提 PR 的 skill 动作尾部自动把 PR 链接和状态回写到表格
   对应行，以便不用手工维护两处状态。
4. 作为 PR 作者，我想在 PR 建好时自动给被指派 reviewer 发企微定向@（带 PR 链接），
   以便 reviewer 第一时间知道，而不是靠群发被淹没。
5. 作为 PR 作者，我想跑 `/nudge` 时自动检测自己名下所有等 review 的 PR，对超过
   阈值无响应的自动再发企微定向@，以便不用手工记账谁欠我 review。
6. 作为 PR 作者，我想可选地开启定时 routine 自动跑 nudge 检测，以便人不在工位
   时催办也在进行。
7. 作为 PR 作者，我想在按 reviewer 意见修改并推送后，用同一个 `/nudge` 通知
   reviewer 再审，以便循环不因"改完没人知道"而卡住。
8. 作为 reviewer，我想收到的提醒是定向@我本人且带 PR 链接的，以便和群闲聊区分、
   点开即达。
9. 作为 reviewer，我想在本机对被指派的 PR 跑 AI review（复用 review skill 的
   full review 流程），意见按 must-fix/should/nit 分级发成 PR comment，
   以便作者知道哪些是阻塞项。
10. 作为 reviewer，我想在我的 AI 判定有 must-fix 时提交 GitHub request-changes，
    以便循环状态机器可查。
11. 作为 reviewer，我想在我的 AI 判定无 must-fix 时收到一份预读简报（类型分派
    叙事 + 风险热图 + AI 自报验证不了的点）作为 PR comment，以便我把有限精力
    投到 load-bearing 行和 AI 盲区，而不是通读全 diff。
12. 作为 reviewer，我想亲手提交 approve（AI 不代按），以便人工把关是真实发生的。
13. 作为 PR 作者，我想在 PR 合入时表格对应行自动置为已完成开发，以便进度实时。
14. 作为团队任何成员，我想跑 `/standup` 看到自己名下待 review 的 PR 和自己
    被指派的 review 任务清单，以便早上开工一眼知道欠账，作为防漏第二道防线。
15. 作为 milestone 负责人，我想在 milestone 各子需求全部合入后在本机跑 `/gate`
    执行完整集成测试 + SIT，以便发版质量有门禁。
16. 作为 milestone 负责人，我想门禁通过时表格自动勾选该 milestone 及其子需求
    完成，以便不用手工核对。
17. 作为 milestone 负责人，我想门禁失败时自动在表格建缺陷子需求行（挂在该
    milestone 下）并企微@归因到的 PR 作者，归因不出来就@我自己，以便失败
    有人认领而不是躺在日志里。
18. 作为新加入的开发，我想 clone harness 仓库跑 install 脚本 + setup skill，
    按模板完成腾讯文档授权和企微 webhook key 配置，以便半小时内接入工作流。
19. 作为团队任何成员，我想我对表格的所有写操作以我自己的腾讯文档身份进行，
    以便表格修改记录自带真实署名。
20. 作为 harness 的维护者，我想 skills 的改进走内部仓库的 PR review，升级对
    全员即 `git pull`，以便工具本身不漂移。
21. 作为 harness 的维护者，我想所有外部副作用都经 `fbh` CLI 且支持 dry-run，
    以便改动可以在不打扰真实表格/企微群的情况下验证。
22. 作为 PR 作者，我想绕过 harness 手工操作 GitHub 之后，下次任何人跑
    standup/nudge 时表格能被顺带对账兜底（后续增强，见 Out of Scope 边界），
    以便镜像滞后不会永久化。

## Implementation Decisions

- **形态与分发**（ADR-0001, 0007）：harness = 独立内部 git 仓库中的一组
  Claude Code skills + 一个 helper CLI + install 脚本；软链到 `~/.claude/skills/`；
  无中心节点、无常驻服务、无 repo secrets。
- **唯一新接缝 = `fbh` CLI**：skills 不直接调 curl/腾讯 API/企微 webhook，
  全部经 `fbh` 子命令。GitHub 查询也统一收口（内部实现可用 gh）。CLI 提供
  `--dry-run`（打印将执行的动作与请求体，不发出）。子命令面（最小集）：
  `fbh sheet upsert-row` / `fbh sheet set-status` / `fbh sheet create-defect-row` /
  `fbh wecom nudge` / `fbh gh review-state` / `fbh gh my-prs` / `fbh gh my-reviews`。
- **skill 命令面**（最小集）：`/setup`（凭证与表格 ID 配置引导）、`/split`
  （总需求→子需求行）、`/claim`、`/pr`（建 PR + 回写 + 首次通知 reviewer）、
  `/nudge`、`/review-pr`（reviewer 侧，复用 review skill，结论落 GitHub review
  状态，终局发预读简报）、`/standup`、`/gate`。
- **真相源与同步**（ADR-0002）：智能表格为主，GitHub 为镜像；动作嵌入式同步——
  状态变更者的 skill 尾部调 `fbh sheet ...` 回写；无轮询无 webhook。
- **催办**（ADR-0001）：作者侧驱动；`/nudge` 幂等（同一 PR 同一天不重复@），
  阈值默认 24h 无 review 活动，可配置。
- **review 循环终止判定**（ADR-0003）：只认被指派 reviewer 的 GitHub review
  状态；作者自己的 AI 自审是提 PR 前的门槛，不计入判定。
- **预读简报**（ADR-0005）：复用 review skill Stage 0 预读简报格式；必含
  AI 自报验证不了的点；作为 PR comment 留档。
- **门禁**（ADR-0004）：`/gate` 包装 fisco-bcos-release-gate 既有工具，不重造；
  归因规则 = 失败场景→涉及模块→该 milestone 内改过该模块的 PR 的作者，
  兜底@milestone 负责人。
- **表格 schema**（本 spec 定案）：子需求表列 = 需求名 / 所属总需求 /
  milestone / 认领人 / 状态 / PR 链接 / 备注；milestone 表列 = 名称 / 负责人 /
  门禁状态 / 门禁时间；缺陷行复用子需求表（备注标 `gate-defect` 来源）。
  子需求状态枚举 = 待认领 → 开发中 → 待review → review循环 → 人工review →
  已合入 → 完成（门禁后）。
- **凭证**（ADR-0006）：模板引导各自本机配置；腾讯文档个人身份授权；
  企微 webhook key 团队共享值各自填入；凭证文件不入任何仓库。

## Testing Decisions

- **好测试的标准**：只断言外部行为——给定状态输入，`fbh --dry-run` 输出的
  动作序列/请求体是否正确；不测 prompt 措辞、不测 CLI 内部函数结构。
- **被测对象**：`fbh` CLI 的全部子命令（唯一接缝）。三类用例：
  (a) dry-run 动作序列断言（如：PR 挂 25h → nudge 输出恰一条企微请求，
  同日第二次跑 → 零输出，验证幂等）；
  (b) fake endpoint 请求体断言（本地起 fake HTTP 服务收请求，断言表格
  upsert 的行字段、企微@的 mention 列表）；
  (c) GitHub 查询解析用例（review-state 对 approve/request-changes/无 review
  三态的判定，用录制的 gh JSON 固件回放）。
- **不测的**：skill prompt 本身（人工验收 + 试用期迭代，eval 明确不做——
  grilling 已确认只选单一 CLI 接缝）；release-gate 工具自带 tests，不重复测。
- **先例参考**：fisco-bcos-release-gate 仓库的 Go 工具 + tests 布局
  （cmd/internal/tests 结构）可作为 `fbh` 的结构范本。

## Out of Scope

- **对账兜底的完整实现**：动作嵌入式同步为主；standup/nudge 顺带全量对账
  先不做，观察镜像滞后严重程度后再立项（ADR-0002 已预留）。
- **skill prompt 的自动化 eval**（claude plugin eval 等）。
- **中心化任何东西**：团队服务器、GitHub Actions 定时扫描、共享密钥库。
- **合入即触发的轻量门禁**：grilling 已选 milestone 手动触发，PR 级快速子集不做。
- **6 人以上/多小队的分层看板**：当前 3-5 人同活拆分，表格单层结构够用。
- **FISCO-BCOS 开源仓库内的任何改动**：harness 与目标仓库零耦合。

## Further Notes

- 催办升级链（超时后 私提醒→群@→@负责人）在 grilling 中未选为必须项，
  `/nudge` 先做单级定向@ + 幂等；升级链留作 `fbh wecom nudge` 的自然扩展位。
- review skill 需要随 harness 仓库分发团队定制版（当前它在个人
  `~/.cc-switch/skills/` 下），注意与个人版共存时的命名冲突。
- 智能表格的"自动提醒"（表格原生@）与企微 bot @ 并用，前者由表格自身配置，
  不经 harness 代码路径。
- 实现偏差（2026-08-20 终审确认）：`fbh sheet create-defect-row` 未做独立
  子命令，由通用 `upsert-row` 覆盖（gate 失败路径内部走 upsert）；review
  结论提交（gh pr review）由 /fbh-review-pr 直调 gh，是单一接缝的已知豁免；
  新增 `mention_map` 配置解决 GitHub 登录名→企微 userid 映射。
