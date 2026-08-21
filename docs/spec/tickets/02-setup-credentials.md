# 02 — /setup 凭证配置引导

**What to build:** 新成员跑 `/setup`，按模板引导完成腾讯文档个人身份授权、填入团队
共享的企微 bot webhook key 和目标智能表格标识；配置落本机、不入任何 git 仓库。
配置完成后 `fbh sheet ping` 以本人身份真实读到表格，证明链路通。（ADR-0006）

**Blocked by:** 01 — 仓库骨架 + install 脚本 + fbh CLI 骨架

**Status:** done (commit d85cc52; 真实腾讯授权+署名终验待用户)

- [x] `/setup` 引导覆盖三项：腾讯文档授权、企微 webhook key、表格标识；缺项时给出可操作的下一步指引
- [x] 凭证文件被 harness 仓库的忽略规则排除，`git status` 验证不可见
- [x] `fbh sheet ping` 成功读取表格并回显表格名；未配置时报出指向 `/setup` 的明确错误
- [x] 腾讯文档操作以配置者本人账号身份进行（表格修改记录署名验证）
