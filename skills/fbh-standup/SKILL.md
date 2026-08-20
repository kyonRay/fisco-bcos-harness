---
name: fbh-standup
description: 防漏看板：一屏列出 我名下等 review 的 PR（挂了多久/催过几次）和 我欠别人的 review（四态循环状态）。早上开工、想知道欠账、检查有没有漏掉的 review 时使用。
---

# /fbh-standup — 防漏看板

```bash
fbh standup
```

输出两节：

1. **我的待 review PR** — url / 三态 / 挂了多久 / 催过几次。
   挂太久的提示用户跑 /fbh-nudge。
2. **我欠的 review** — url + 循环状态：
   - `等我首审` → 提示用户跑 /fbh-review-pr
   - `等作者改` → 不用动，球在作者那
   - `等我复审` → 作者已改，提示用户跑 /fbh-review-pr 复审
   - `等人工approve` → 预读简报已发，提示用户亲自看 load-bearing 行然后 approve

两节皆空输出"无欠账 ✅"。

## 用法建议

把输出念给用户时按紧急度排序：等我复审/等我首审 在前（阻塞别人），
自己的 PR 催办在后。不替用户执行 review 或 approve——看板只指路。
