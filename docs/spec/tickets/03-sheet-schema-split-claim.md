# 03 — 表格 schema + /split 拆分 + /claim 认领

**What to build:** 需求负责人跑 `/split` 把一项总需求拆成多个子需求行写进智能表格
（真相源），列含：需求名 / 所属总需求 / milestone / 认领人 / 状态 / PR 链接 / 备注；
状态枚举 = 待认领 → 开发中 → 待review → review循环 → 人工review → 已合入 → 完成。
开发跑 `/claim` 认领某行，认领人与状态即时回写（动作嵌入式同步）。
milestone 表含：名称 / 负责人 / 门禁状态 / 门禁时间。（ADR-0002，spec 表格 schema 定案）

**Blocked by:** 02 — /setup 凭证配置引导

**Status:** done (commit e595542; 真实表格终验待用户)

- [x] `/split` 在表格创建符合 schema 的子需求行；`/claim` 置认领人并推进状态到"开发中"
- [x] `fbh sheet upsert-row` / `fbh sheet set-status` 的 dry-run 测试断言行字段与状态枚举合法性
- [x] fake endpoint 测试断言真实请求体的列名与取值
- [x] 已被认领的行再次 `/claim` 被拒绝并提示当前认领人
