---
name: fbh-review-pr
description: reviewer 侧 AI review 循环：对被指派的 PR 做完整 review，意见按 must-fix/should/nit 分级发 PR comment；有 must-fix 提交 request-changes，无 must-fix 发预读简报等人工 approve。用户被指派 review、说"帮我审这个 PR"、收到催办要处理 review 时使用。
---

# /fbh-review-pr — reviewer 侧 AI review 循环

终止判定以本 skill 的产出为准（ADR-0003）：结论必须落 GitHub 原生 review
状态，绝不停留在会话里。**approve 永远由 reviewer 本人按，AI 不代按。**

## 流程

### 1. 审（每轮都完整做）

对 PR 做完整 review，要求：

- **开场叙事**：先分类 PR（修复/新增特性/重构），按类型讲清
  问题→触发条件→根因→修法（修复类）或 文件角色→入口→旧代码如何用新代码
  （特性类）或 旧形态→新形态→等价性（重构类），再谈发现。
- **逐条验证**：每个 finding 对着 PR 实际 committed blob 确认代码路径后才定级，
  验证不了的明说是猜测。检查循环上一轮的意见是否真的修了（看代码不看回复）。
- **分级**：每条意见标 `[must-fix]` / `[should]` / `[nit]`。只有 must-fix
  阻塞循环终止。分级从严——不确定是不是 must-fix 的降为 should 并说明疑点。

### 2. 出结论（二选一，都要落 GitHub 状态）

**有 must-fix** → 把分级意见发成 PR review：

```bash
gh pr review <PR> --request-changes --body "<分级意见全文>"
```

然后跑 `fbh sheet set-status --table <需求表> --key "<需求名>" --status review循环`
（首轮时；已是该状态则跳过）。作者改完会用 /fbh-nudge 催复审。

**无 must-fix** → 发预读简报，进入人工阶段：

```bash
gh pr review <PR> --comment --body "<预读简报>"
```

简报**必含三要素**，缺一不可：

1. **类型分派叙事**（同上，最终版）。
2. **风险热图**：把 diff 拆成 mechanical（skim 即可，注明"我已查过一致性"）
   与 load-bearing（并发/生命周期/共识/存储/密码学/序列化），load-bearing
   条目按 `path:line` 排序、每条一句为什么险。
3. **AI 自报验证不了的点**：需要业务/领域知识的判断、未能实证的并发时序假设、
   测试未覆盖的路径。这是人工不可替代的部分，绝不许写"无"敷衍。

发完简报跑 `fbh sheet set-status ... --status 人工review`，并告诉 reviewer：
"简报已发，请重点看热图中 load-bearing 条目和第 3 节盲区，看完自行 approve。"

### 3. 人工 approve 之后（reviewer 本人按完 approve 再回来）

```bash
fbh gh review-state --pr <PR>        # 确认已 approved
```

合入由作者或 reviewer 按团队惯例执行；合入后作者跑
`fbh sheet set-status ... --status 已合入`。

## 与个人 review skill 的关系

本 skill 是团队定制版，独立命名 `fbh-review-pr`，与任何个人 `review` skill
共存不冲突；个人版有更细的流程时可用个人版完成第 1 步，但第 2 步的
GitHub 状态落地与表格回写必须按本 skill 执行。
