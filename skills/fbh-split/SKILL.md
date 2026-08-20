---
name: fbh-split
description: 把一项总需求拆成多个子需求行写进团队腾讯智能表格（真相源）。需求负责人做需求拆分、立项、建 milestone 时使用。
---

# /fbh-split — 总需求拆分

把用户描述的总需求拆成可独立交付的子需求，逐行写入团队智能表格的需求表。

## 流程

1. **先看现状**：`fbh sheet ping` 列出工作表，确认需求表的 sheet_id
   （首次使用时记录下来，后续直接复用）。
2. **和用户一起拆分**：每个子需求应可独立开发、独立提 PR。拆分粒度和
   milestone 归属由用户拍板，不要替用户决定。
3. **逐行写入**（一行一条）：

   ```bash
   fbh sheet upsert-row --table <需求表sheet_id> --key "<子需求名>" \
     --set "所属总需求=<总需求名>" \
     --set "milestone=<M名>" \
     --set "状态=待认领"
   ```

   需求名是行的唯一键：同名 upsert 会更新既有行而不是重复建行。
4. **milestone 行**（如是新 milestone）写入 milestone 表：

   ```bash
   fbh sheet upsert-row --table <milestone表sheet_id> --key-field 名称 \
     --key "<M名>" --set "负责人=<负责人>"
   ```

5. **回显清单**：把写入的行列出来给用户核对。

## 状态枚举（固定，写错会被 CLI 拒绝）

待认领 → 开发中 → 待review → review循环 → 人工review → 已合入 → 完成

## 注意

- 先 `--dry-run` 一遍给用户看动作，确认后去掉 dry-run 真写。
- 表格是唯一真相源（ADR-0002）：拆分结果只写表格，不建 GitHub issue。
