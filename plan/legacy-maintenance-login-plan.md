# 历史维护与登录计划

本文件用于保留旧的根目录 `plan.md` 中对后续开发有价值的上下文，但不会替代原文件。原始 `plan.md` 仍保留在仓库根目录。

## 已完成或需要保留的上下文

- 后端同步流程会先拉取 Replive chat rooms，再按 room 分页拉取 chat messages。
- 媒体消息会下载到本地用于归档，但前端聊天展示不应通过本地服务器读取这些文件，避免带宽成本。
- `saveMessage` 已从“整批媒体全部下载成功后再入库”改成“单条消息下载成功后立即入库”，避免后续下载超时时前面记录丢失。
- SQLite driver 已切换为 GORM 兼容的纯 Go driver，默认构建不需要 CGO。
- 主程序增加了 panic recover，panic 信息和 stack trace 会写入 `replive_*.log`；Windows 双击运行时会暂停窗口，避免报错信息一闪而过。
- 已实现重复 refresh token `401` 的处理：多次失败后归档旧 token，并清空当前配置 token，使下次启动重新登录。
- 已实现 Twitter 登录流程：PKCE、`GetSNSLoginURL`、`replive-user-auth://user-auth` 回调解析、`UserAuthBySNS`。

## 需要保留的登录行为

- Google 登录继续使用现有 OAuth 流程。
- Twitter 登录流程使用：
  - `GetSNSLoginURLRequest{id_provider=TWITTER, state="", code_challenge=<PKCE challenge>}`
  - 浏览器授权
  - 回调参数 `oauth_token` 和 `oauth_verifier`
  - `UserAuthBySNSRequest{id_provider=TWITTER, oauth_token, oauth_verifier, code_verifier}`
- 如果已有 Replive refresh token 多次刷新失败并返回 `401`，先归档旧 token，再清空当前配置中的 token。
- 下次启动时，因为当前 `refresh_token` 为空，程序会重新进入配置的登录方式。

## 需要保留的后端稳定性规则

- 不要在持有 DB 写锁时执行网络请求或重文件系统操作。
- 分页循环必须防御空游标、游标不前进、游标重复和最大页数。
- worker 不能因为 nil channel 或满 channel 永久阻塞。
- 构建和测试应继续支持 `CGO_ENABLED=0`。

## 历史计划位置

- 原始文件：`plan.md`
- 当前新功能计划：`plan/chat-ui-multi-room-cursor-translation.md`
