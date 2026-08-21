# fbh 团队使用文档

fbh 是团队工作流 harness：一套装在每个开发本机的 Claude Code skills +
`fbh` CLI，无中心服务。通知采用**半自动播报**：fbh 起草现成的企微群
文案（PR 链接、状态、还差谁 approve、台账链接），作者复制粘贴进群——
本团队的企微禁用了群机器人、个人版表格自动化又到不了企微群，"人来
送达"是零权限依赖的唯一通道。台账（腾讯智能表格）做全队共享状态。
按需两档接入，随时升档、命令不变：

| 档位 | 解决什么 | 需要配什么 |
|---|---|---|
| ① PR 台账（**入门档**） | 催 review 靠吼、循环没终点、谁欠 review 看不见 | 腾讯授权 + 台账工作表 |
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

1. 智能表格里建"PR台账"工作表，8 个**文本列**，列名完全一致：
   `PR链接 | 标题 | 作者 | reviewers | 已approve | 状态 | 待处理人 | 更新时间`
   （openapi 建不了人员列也写不了成员值，全文本即可——播报文案里
   直接写 GitHub 用户名）。
2. 把表格分享给全队（可编辑）；`sheet_file_id` 用表格 URL 的尾段配，
   播报文案里的台账链接就是由它拼出来的。

**② 完整档补充：** 再建需求表（7 列：需求名/所属总需求/milestone/
认领人/状态/PR链接/备注）和 milestone 表（4 列：名称/负责人/门禁状态/
门禁时间）；milestone 负责人另配 `gate_cmd`。

## 三、日常流程

### 提 PR（作者）

`/fbh-pr` —— 一条命令：建 GitHub PR 指派 reviewer（可多人，逗号分隔）→
**自动在台账注册一行**（状态=待review，待处理人=全部 reviewer）→ 输出
一段播报文案，**复制粘进企微群**。完整档加 `--table/--key` 可同时挂需求行。

### 按 review 意见修复后（作者）

改完 push，然后：

```bash
fbh pr sync --pr <PR链接>
```

状态从 GitHub 自动推导写回台账（谁 approve 了记入"已approve"列），
**"待处理人"刷成还没 approve 的人**，播报文案只点名他们——已 approve
的同事不被打扰。把文案粘进群即完成"催复审"。

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

再跑一次 `fbh pr sync --pr <链接>`：状态和待处理人重新核对后生成新的
播报文案，粘群里再催一轮。文案永远只点名还没 approve 的人。

### 合入后（作者）

`fbh pr sync --pr <PR链接>` 把台账行刷成"已合入"、待处理人清空
（不再生成播报）。

### 完整档补充

需求负责人 `/fbh-split` 拆需求、开发 `/fbh-claim` 认领、milestone 负责人
`/fbh-gate` 跑门禁（通过批量勾选完成，失败自动建缺陷行、认领人=归因作者，
负责人把结果转达群里）。详见各 skill。

## 四、命令参考

| 命令 | 作用 |
|---|---|
| `fbh config set/show` | 本机配置（sheet_file_id/pr_sheet_id/mcp_base_url/gate_cmd） |
| `fbh sheet ping` | 连通性验证，列工作表 |
| `fbh pr open` | 建 PR → 台账注册 → [需求行回写] → 出播报文案 |
| `fbh pr sync --pr <url>` | 从 GitHub 刷台账状态与待处理人 → 出播报文案 |
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
| 播报里没有台账链接 | `sheet_file_id` 配的不是表格 URL 尾段 |
| `pr sync` 报 gh 错误 | PR 链接写错或 `gh auth status` 失效 |
| 台账出现重复行 | 两人同时首次 sync 同一 PR 的竞态，删掉一行即可（后续 sync 幂等） |
| `already claimed by X` | 需求已被 X 认领（完整档），找 X 协调 |
| pr open 中途失败 | 跑 `fbh pr sync --pr <链接>` 补齐台账行，别重跑造重复 PR |

## 七、设计边界（为什么是这样）

- 无中心服务：状态推进动力在作者侧——谁的 PR 谁着急（ADR-0001）。
- **fbh 起草、人来送达**：本团队企微禁群机器人，个人版表格自动化又
  发不进企微群，自动送达无路可走；fbh 把"组织语言"这步做掉（点名
  永远来自"待处理人"列，从 GitHub 推导），作者一次粘贴完成送达。
- **真相源分层**：review 状态的真相源永远是 GitHub 原生 review 状态
  （ADR-0003），台账是全队可见的镜像，`pr sync` 单向刷入；需求/milestone
  状态的真相源才是表格（ADR-0002，完整档）。所以台账状态不用手改，
  改了也会被下次 sync 覆盖。
- 凭证永不入库；腾讯身份是个人的（ADR-0006）。
- 全部外部副作用走 fbh 单一接缝，dry-run/fake-endpoint 可测。已知豁免：
  review 结论提交（`gh pr review`）由 /fbh-review-pr 直调 gh——它天然要
  人参与、且 GitHub 原生状态即真相源，不经 fbh 包装。
