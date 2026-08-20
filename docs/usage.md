# fbh 团队使用文档

fbh 是团队工作流 harness：一套装在每个开发本机的 Claude Code skills +
`fbh` CLI，无中心服务。**fbh 只写表格，不直发任何消息**——所有通知由
腾讯智能表格的自动化规则发到企微，谁该被提醒由 fbh 维护的"待处理人"
列决定。按需两档接入，随时升档、命令不变：

| 档位 | 解决什么 | 需要配什么 |
|---|---|---|
| ① PR 台账（**入门档**） | 催 review 靠吼、循环没终点、谁欠 review 看不见 | 腾讯授权 + 台账工作表 + 3 条表格自动化 |
| ② 完整 | ① + 需求拆分/认领/milestone 门禁 | ① + 需求表 + milestone 表 |

## 一、安装（每人一次）

```bash
git clone https://github.com/kyonRay/fisco-bcos-harness.git && cd fisco-bcos-harness
./install.sh     # 软链全部 skills 到 ~/.claude/skills/，构建 bin/fbh
```

把 `bin/fbh` 加进 PATH（如 `ln -s $PWD/bin/fbh /opt/homebrew/bin/fbh`）。
升级：`git pull && ./install.sh`。之后在 Claude Code 里跑 `/fbh-setup` 逐项引导。

## 二、配置

### 每人本机（5 分钟）

1. 腾讯文档授权（/fbh-setup 引导，浏览器扫码一次；token 由 mcporter 保管，
   fbh 复用，表格操作带你本人署名）。
2. ```bash
   fbh config set sheet_file_id '<表格URL里 smartsheet/ 后面那段>'
   fbh config set pr_sheet_id  '<台账工作表 sheet_id>'   # fbh sheet ping 可查
   ```

验证：`fbh sheet ping` 列出工作表名即链路通。

### 团队一次性（负责人做）

1. 智能表格里建"PR台账"工作表，8 列，列名完全一致：
   `PR链接 | 标题 | 作者 | reviewers | 已approve | 状态 | 待处理人 | 更新时间`
   —— **待处理人 建为人员列**（自动化按它定向提醒），其余文本列。
2. 表格右上角"自动化"面板配 3 条规则（通知全靠它们）：
   - **添加记录提醒到群**：台账新增记录 → 发消息到团队群（新 PR 通知）；
   - **修改后提醒负责人**：台账记录变更 → 发文档消息给 [待处理人]
     （能加条件则限定 状态=待复审，更精准也省配额）；
   - **定时提醒负责人**（可选）：每天定时给 [待处理人] 发文档消息（周期催办）。
3. 注意自动化每月运行次数配额（免费 100 次/月）；团队 PR 流量大时
   找管理员提额。

**② 完整档补充：** 再建需求表（7 列：需求名/所属总需求/milestone/
认领人/状态/PR链接/备注）和 milestone 表（4 列：名称/负责人/门禁状态/
门禁时间）；milestone 负责人另配 `gate_cmd`。

## 三、日常流程

### 提 PR（作者）

`/fbh-pr` —— 一条命令：建 GitHub PR 指派 reviewer（可多人，逗号分隔）→
**自动在台账注册一行**（状态=待review，待处理人=全部 reviewer）→ 表格
自动化把新 PR 推到群里。完整档加 `--table/--key` 可同时挂需求行。

### 按 review 意见修复后（作者）

改完 push，然后：

```bash
fbh pr sync --pr <PR链接>
```

状态从 GitHub 自动推导写回台账（谁 approve 了记入"已approve"列），
**"待处理人"刷成还没 approve 的人**——表格自动化只提醒他们，已 approve
的同事不被打扰。

### 看全队欠账（任何人，问自己的 AI 即可）

- `fbh pr board` —— 台账里所有未完结 PR：链接/状态/reviewers/谁已 approve。
- `/fbh-standup` —— 个人视角两节：我的 PR 等谁（挂了多久）+
  我欠谁的（等我首审/等作者改/等我复审/等人工approve）。

### 被指派 review（reviewer）

`/fbh-review-pr` —— AI 完整 review，意见按 must-fix/should/nit 分级发 PR
comment：有 must-fix → request-changes 进循环；无 must-fix → 发**预读简报**
（叙事 + 风险热图 + AI 自报盲区），人照简报精读 load-bearing 行后
**亲手 approve**（AI 永不代按）。

### 没人理（作者）

不需要手动催：表格的**定时提醒自动化**会周期性提醒"待处理人"列里的人。
想立刻再提醒一次，跑 `fbh pr sync --pr <链接>`——记录变更会再触发一次
"修改后提醒负责人"。

### 合入后（作者）

`fbh pr sync --pr <PR链接>` 把台账行刷成"已合入"、待处理人清空
（此后任何自动化都不再提醒任何人）。

### 完整档补充

需求负责人 `/fbh-split` 拆需求、开发 `/fbh-claim` 认领、milestone 负责人
`/fbh-gate` 跑门禁（通过批量勾选完成，失败自动建缺陷行、认领人=归因作者，
表格自动化提醒到人）。详见各 skill。

## 四、命令参考

| 命令 | 作用 |
|---|---|
| `fbh config set/show` | 本机配置（sheet_file_id/pr_sheet_id/mcp_base_url/gate_cmd） |
| `fbh sheet ping` | 连通性验证，列工作表 |
| `fbh pr open` | 建 PR → 台账注册 → [需求行回写] |
| `fbh pr sync --pr <url>` | 从 GitHub 刷台账状态与待处理人 |
| `fbh pr board` | 台账中所有未完结 PR |
| `fbh gh review-state / my-prs / my-reviews` | PR review 状态与欠账查询 |
| `fbh standup` | 个人防漏看板 |
| `fbh sheet upsert-row / set-status / claim` | 需求表行操作（完整档） |
| `fbh gate run` | milestone 门禁（完整档） |

**任何子命令加 `--dry-run`**：外部动作只打印 JSON 行不执行。skills 真写前
都会先 dry-run 给你确认。例外：`fbh gate run --dry-run` 的门禁命令本身
仍会真跑（可能数小时），dry 的只是表格回写。

## 五、状态机

**PR 台账（由 GitHub 状态自动推导，永远不用手填）：**

```
待review → 修复中 → 待复审 → 已approve → 已合入
 (pr open)  (被request- (修复已   (全员      (合入,
             changes)    推送)    approve)   pr sync 刷入)
```

"待处理人"列同步维护：待review=全部 reviewer；待复审=未 approve 的人；
已approve/已合入=空。

**需求表（完整档，CLI 强校验）：**

```
待认领 → 开发中 → 待review → review循环 → 人工review → 已合入 → 完成
 (split)  (claim)  (pr open)  (有must-fix)  (简报发出)  (手动置)  (gate过)
```

## 六、故障排查

| 症状 | 处理 |
|---|---|
| `not configured; run /fbh-setup` | 跑 `/fbh-setup` 补缺项（报错会说缺哪个 key） |
| `no tencent-docs token` | 腾讯授权没走完，/fbh-setup 引导重走 |
| MCP HTTP 401/403 | token 过期，重新授权 |
| 落表成功但没人收到提醒 | 表格自动化没配/被停用，或本月配额用完（自动化面板可查） |
| 提醒发了但 @ 不到人 | 待处理人列不是人员列，或列里的名字匹配不到表格成员 |
| `pr sync` 报 gh 错误 | PR 链接写错或 `gh auth status` 失效 |
| 台账出现重复行 | 两人同时首次 sync 同一 PR 的竞态，删掉一行即可（后续 sync 幂等） |
| `already claimed by X` | 需求已被 X 认领（完整档），找 X 协调 |
| pr open 中途失败 | 跑 `fbh pr sync --pr <链接>` 补齐台账行，别重跑造重复 PR |

## 七、设计边界（为什么是这样）

- 无中心服务：状态推进动力在作者侧——谁的 PR 谁着急（ADR-0001）。
- **通知与状态合一**：fbh 只写状态，通知是状态变更在表格侧的副作用
  （自动化规则），不存在第二条并行通知通道。谁被提醒 = "待处理人"列，
  由 `pr sync` 从 GitHub 推导。
- **真相源分层**：review 状态的真相源永远是 GitHub 原生 review 状态
  （ADR-0003），台账是全队可见的镜像，`pr sync` 单向刷入；需求/milestone
  状态的真相源才是表格（ADR-0002，完整档）。所以台账状态不用手改，
  改了也会被下次 sync 覆盖。
- 凭证永不入库；腾讯身份是个人的（ADR-0006）。
- 全部外部副作用走 fbh 单一接缝，dry-run/fake-endpoint 可测。已知豁免：
  review 结论提交（`gh pr review`）由 /fbh-review-pr 直调 gh——它天然要
  人参与、且 GitHub 原生状态即真相源，不经 fbh 包装。
