# fbh 团队使用文档

fbh 是团队工作流 harness：需求拆分 → 认领 → 开发 → PR → AI review 循环 →
人工 review → milestone 门禁 → 勾选完成，全流程跑在每个开发本机的
Claude Code skills + `fbh` CLI 上，无中心服务。腾讯智能表格是需求状态的
唯一真相源，GitHub 是镜像。

## 零、最小接入：只治 PR review（推荐起步，5 分钟/人）

不动需求表、不碰 milestone、不需要腾讯文档授权——只解决"催 review 靠吼、
循环没终点、人工 review 走过场"：

```bash
git clone https://github.com/kyonRay/fisco-bcos-harness.git && cd fisco-bcos-harness
./install.sh
fbh config set wecom_webhook '<企微群机器人 webhook URL>'
fbh config set mention_map '{"<github登录名>":"<企微userid>", ...}'
```

日常只用四个动作：

| 场景 | 命令/skill |
|---|---|
| 提 PR | `/fbh-pr`（不带表格参数 → 建 PR + 企微定向@ 两连） |
| 没人理 | `/fbh-nudge`（超 24h 才催、同日幂等、不骚扰在等作者的） |
| 被指派 review | `/fbh-review-pr`（must-fix 分级 → request-changes 循环 → 预读简报 → 人亲手 approve） |
| 早上开工 | `/fbh-standup`（我欠谁的 / 谁欠我的，四态一目了然） |

表格那半（/fbh-split、/fbh-claim、/fbh-gate、状态回写）想接的时候再走
下面的完整配置，命令不变，加上 `--table/--key` 即自动升级为三连。

## 一、完整安装（每人一次，约 10 分钟）

```bash
git clone <内部仓库地址> && cd fisco-bcos-harness
./install.sh          # 软链全部 skills（共 9 个）到 ~/.claude/skills/，构建 bin/fbh
```

把 `bin/fbh` 加进 PATH（或 `ln -s $PWD/bin/fbh /usr/local/bin/fbh`）。
升级：`git pull && ./install.sh`（skills 即时生效，fbh 重新构建）。

然后在 Claude Code 里跑 `/fbh-setup`，按引导完成三项配置：

1. **腾讯文档授权**——走 tencent-docs 的浏览器授权流程，token 由 mcporter
   保管，fbh 直接复用，以你本人身份操作表格（修改记录带真实署名）。
2. **企微 webhook key**——团队共享值，找负责人要。
3. **表格 file_id**——团队智能表格 URL 里 `smartsheet/` 后面那段。
4. （可选但强烈建议）**企微身份映射**——GitHub 用户名和企微 userid 不同名时
   @会落空：`fbh config set mention_map '{"github登录名":"企微userid", ...}'`。

验证：`fbh sheet ping` 列出工作表名即为链路通。

## 二、日常流程（按角色）

### 需求负责人：拆需求

`/fbh-split` —— 把总需求拆成子需求行写进需求表（状态=待认领），
新 milestone 顺带写 milestone 表。

### 开发：认领 → 开发 → 提 PR

1. `/fbh-claim` —— 认领一条子需求（已被人认领会拒绝并告知认领人）。
2. 正常开发、push 分支。
3. `/fbh-pr` —— 一条命令：建 PR 指派 reviewer → 表格回写 PR链接+待review →
   企微定向@ reviewer。
4. 没人理？`/fbh-nudge` —— 只催超过阈值（默认 24h）无响应的，同一 PR
   同一天最多催一次；改完代码想催复审用 `--threshold 1`。

### reviewer：AI review 循环 → 人工把关

1. 企微收到@后，`/fbh-review-pr` —— AI 完整 review，意见按
   must-fix/should/nit 分级发 PR comment：
   - 有 must-fix → request-changes，等作者改（作者会再催）。
   - 无 must-fix → 发**预读简报**（叙事 + 风险热图 + AI 自报盲区）。
2. 人工阶段：照简报只精读 load-bearing 行和 AI 盲区，**亲手 approve**
   （AI 永不代按）。
3. 合入后作者把表格状态推到"已合入"。

### 每个人：早上开工

`/fbh-standup` —— 两节看板：我的 PR 等谁 review（挂了多久/催过几次）+
我欠谁的 review（等我首审/等作者改/等我复审/等人工approve）。
两节皆空 = 无欠账。

### milestone 负责人：门禁

一次性配置 `fbh config set gate_cmd '<release-gate 调用命令>'`，然后
`/fbh-gate` —— 跑完整集成测试+SIT：
- 通过 → milestone 行记"通过"+时间，该 milestone 全部子需求置"完成"。
- 失败 → 自动建缺陷行（含失败输出尾部）+ 企微@归因作者（AI 从失败输出
  和 PR 记录里归因；归因不出@负责人本人）。

## 三、命令参考

| 命令 | 作用 |
|---|---|
| `fbh config set/show` | 本机配置（show 对 webhook 打码） |
| `fbh sheet ping` | 连通性验证，列工作表 |
| `fbh sheet upsert-row / set-status / claim` | 表格行操作（状态枚举强校验） |
| `fbh pr open` | 建 PR → [回写] → 企微@（--table/--key 可选，不带则跳过表格） |
| `fbh nudge run [--threshold N]` | 幂等催办 |
| `fbh gh review-state / my-prs / my-reviews` | PR review 状态与欠账查询 |
| `fbh standup` | 防漏看板 |
| `fbh gate run` | milestone 门禁 |

**任何子命令加 `--dry-run`**：外部动作只打印 JSON 行不执行。skills 在真写
之前都会先 dry-run 给你确认。注意 `fbh gate run --dry-run` 中**门禁命令本身
仍会真跑**（可能数小时），dry 的只是表格/企微回写。

## 四、状态机（需求表"状态"列，CLI 强校验）

```
待认领 → 开发中 → 待review → review循环 → 人工review → 已合入 → 完成
  (split)  (claim)   (pr open)  (review-pr    (简报发出)   (合入后    (gate
                                 有must-fix)                手动置)    通过)
```

## 五、故障排查

| 症状 | 处理 |
|---|---|
| `not configured; run /fbh-setup` | 跑 `/fbh-setup` 补缺项 |
| `no tencent-docs token` | tencent-docs 授权流程没走完，见 /fbh-setup 第二步 |
| MCP HTTP 401/403 | token 过期，重新授权 |
| `already claimed by X` | 需求已被 X 认领，找 X 协调，别硬抢 |
| pr open 中途失败 | 按报错单独补跑 sheet/wecom 一步，别重跑造重复 PR |
| 催办没发出 | 看是否同日已催过（幂等）；`fbh gh my-prs` 核对 PR 状态 |
| 绕过 fbh 手工改了 GitHub | 表格镜像会滞后，下次相关 skill 动作时补写回 |

## 六、设计边界（为什么是这样）

- 无中心服务：催办动力在作者侧——谁的 PR 谁着急（ADR-0001）。
- 表格为主 GitHub 为镜像，改状态的人顺带写回，无轮询（ADR-0002）。
- review 循环终止只认 reviewer 侧的 GitHub review 状态（ADR-0003）。
- 凭证永不入库；腾讯身份是个人的，企微 key 是共享的（ADR-0006）。
- 全部外部副作用走 fbh 单一接缝，dry-run/fake-endpoint 可测（spec）。
  已知豁免：review 结论提交（`gh pr review`）由 /fbh-review-pr 直调 gh——
  它天然要人参与、且 GitHub 原生状态即真相源，不经 fbh 包装。
