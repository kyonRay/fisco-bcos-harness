---
name: fbh-gate
description: milestone 门禁：在负责人本机跑完整集成测试+SIT（包装 fisco-bcos-release-gate），通过则回写表格勾选完成，失败则自动建缺陷行（表格自动化随即提醒归因作者）。milestone 到点验收、发版前验证时使用。
---

# /fbh-gate — milestone 门禁

只有 milestone 负责人跑（ADR-0004）。门禁命令本体是 fisco-bcos-release-gate
（或团队配置的其它命令），fbh 只包装：执行 → 按结果回写表格；通知由表格的
添加记录提醒自动化发出。

## 前置（一次性）

```bash
fbh config set gate_cmd '<release-gate 的完整调用命令>'
```

## 流程

1. **确认时机**：该 milestone 下子需求应全部"已合入"（`fbh sheet ping` +
   看表确认；有未合入的先提示用户，用户坚持再继续）。
2. **先 dry-run 看回写动作**（门禁命令本身会真跑，外部回写不执行）：

   ```bash
   fbh gate run --milestone <M名> --milestone-table <milestone表id> \
     --reqs-table <需求表id> --owner <负责人> --dry-run
   ```

3. **真跑**（去掉 --dry-run）。耗时可能很长，提醒用户机器会被占用。
4. **结果处理**：
   - **通过**：milestone 行已记 门禁状态=通过+时间，该 milestone 全部子需求
     状态=完成。向用户报喜并复述勾选结果。
   - **失败**：exit 1，表格已自动建缺陷行（gate-defect-<M>-<时间>，状态待认领，
     认领人=归因对象，备注含失败尾部输出）；表格自动化会提醒到人。

## 归因（失败时，AI 的职责）

CLI 不做语义归因——**你来做**：读门禁失败输出，定位失败场景涉及的模块，
查该 milestone 内改过该模块的 PR（`gh pr list --search "milestone:<M>"` 或表格
PR链接列），找出最可能的引入者，然后带 `--assignee <该作者>` 重跑失败回写
（或直接改缺陷行认领人）。归因不出来就不传 --assignee，CLI 默认@负责人
自己分诊（ADR-0004 兜底）。

## 注意

- 失败路径绝不改动既有子需求行状态，只新增缺陷行。
- 重跑门禁前先处理掉上一轮缺陷行（认领或关闭），避免堆积。
