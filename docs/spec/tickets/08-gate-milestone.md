# 08 — /gate milestone 门禁：通过勾选、失败建缺陷行

**What to build:** milestone 负责人在本机跑 `/gate <milestone>`：包装
fisco-bcos-release-gate 既有工具执行完整集成测试 + SIT。通过 → 回写表格勾选该
milestone 及其下全部子需求为"完成"，记录门禁时间。失败 → 自动在表格新建缺陷
子需求行（挂该 milestone 下、备注标 gate-defect），按 失败场景→涉及模块→该
milestone 内改过该模块的 PR 作者 归因并企微定向@；归因不出则@负责人本人。
（ADR-0004）

**Blocked by:** 04 — /pr 全链

**Status:** done (commit 773a3a7; 真实 release-gate 串接终验待用户)

- [x] 通过路径：milestone 行门禁状态/时间与子需求行"完成"状态一次跑批全部回写（dry-run 断言动作序列）
- [x] 失败路径：缺陷行字段符合子需求 schema 且备注含 gate-defect；企微@目标为归因作者，归因失败时为 milestone 负责人
- [x] release-gate 工具以包装方式调用，不复制其逻辑；其自身 tests 不重复
- [x] 门禁失败时 milestone 保持未完成，已有子需求行状态不被误改
