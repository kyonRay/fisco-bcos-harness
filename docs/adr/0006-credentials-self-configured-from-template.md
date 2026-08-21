# ADR-0006: 凭证各自本机配置，setup skill + 模板引导，不入库

Status: accepted
Date: 2026-08-20

## Context

每个开发本机的 skills 都要访问腾讯文档 API（读写智能表格）和企微群 bot webhook
（发定向@）。3-5 人团队，无密钥管理基础设施。

## Decision

1. harness 仓库提供 setup skill + 配置模板（`.env.template` 或等价物）。
2. 每人首次安装时：自己完成腾讯文档授权登录；填入团队共享的企微 bot webhook key。
3. 凭证文件不入任何 git 仓库（harness 仓库 .gitignore 排除）。

## Consequences

- 无共享密钥库依赖；换 webhook key 时需要通知每人手动更新（3-5 人可接受）。
- 腾讯文档操作以各人自己的账号身份进行，表格修改记录自带真实署名。
