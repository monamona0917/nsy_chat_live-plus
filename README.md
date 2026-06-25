# nsy_chat_live

## 坏事都是 codex + gpt 做的，跟我没关系哦。
详细使用说明见飞书：https://my.feishu.cn/wiki/PXe9wkiksifZR9kpVoucsKs1nQe

## 功能说明
1. 登录：谷歌账号登录或 x 登录，跳转到网页完成授权。其他的账号体系未支持，可自行抓包获取 refresh token 后写到 config.yaml 使用。
2. 自动拉取你关注的推的头像和背景图片；拉取并存储你订阅的推的所有消息，保存图片和视频到本地。
3. 监控你关注的推是否直播，使用 ffmpeg 开启录制。
4. 提供 replive-web 网页，比 app 多了个按时间跳转消息和关键词搜索消息的功能，方便快速查找和回忆，总比破 app 上划个半天好用。如果你取消订阅了某个人，app上会清除，但这里会永久保存。


## 开发建议
1. 导出 apk 让 AI 帮你解包，你想干啥就让 AI 帮你基于解包代码去弄就行

### 简要说明
1. 准备环境：go、proto、npm、ffmpeg
2. 接口统一往 proto 写好协议然后生成代码然后调用。
3. 如需调试接口，可以得到 response 后去推导结构，或者直接 base64 转为字符串，然后丢给 ai 分析。或者token够的话直接让AI去解包最方便。

