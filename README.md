# gpt-bookingMeetingRoom
企业微信AI会议室应用，一句话预定会议室

## 前置准备
### 企业微信
1. 创建一个企业自建应用
2. 登录管理后台，给该应用以下权限
   1. 协作 -> 日程 -> 可调用接口的应用
   2. 协作 -> 会议 -> 可调用接口的应用
3. 为所有人分配拼音的企业邮箱
### OpenAI
1. 准备一个可以调用API接口的有效token
### 服务器
1. 服务器需要可以访问OpenAI接口

## 技术方案简述
1. 启用API接受消息：可以接受到用户通过企业微信发送给该应用的消息
2. 自然语言解析，将用户发送的自然语言消息，通过使用OpenAI的能力，提取预定会议室必须的信息并转化为结构化的消息 
3. 根据用户的需求，匹配并预定会议室
   1. 如果用户提供的消息不够，通过问答的方式要求用户补充信息（TODO）
4. 根据匹配和预定的情况，给用户反馈（TODO）

## 操作流程


#### 0.代码说明
*接口*
```
GET    /ping                     服务健康检查
GET    /wechat/check             企业微信服务器验证地址，在企业微信后台配置 域名+/wechat/check
POST   /wechat/check             企业微信服务器事件推送地址地址
```

*配置文件*
参考 `.env.example` 文件，创建一个 .env 文件。服务启动的时候会 load `.env`, 如果不存在会 panic！

参数含义
```
# 验证企业微信回调的token
WEWORK_TOKEN=token
# 验证企业微信回调的key
WEWORK_ENCODING_AEK_KEY=encodingAesKey
# 企业微信企业id
WEWORK_CORP_ID=corpid
# 企业微信secret
WEWORK_CROP_SECRET=corpsecret
# openai key
OPENAI_KEY=key
```

#### 1.登陆（注册）你的 OpenAI 账号，拿到对应的 key
参数会用到 [gpt.go](./service/openai_api.go) 当中

#### 2.注册并登陆企业微信后台
应用管理 - 自建应用

#### 3.配置应用服务器
host + `/wechat/check`

注意，只有这些参数和企业微信`接收事件服务器`一致的时候，才能验证通过。代码中的 corpsecret 一定是通过企业微信获得的，首次获取一定是`企业微信app发送`
![](https://raw.githubusercontent.com/razertory/statics/main/staic/4.png)
![](https://raw.githubusercontent.com/razertory/statics/main/staic/5.png)


## 其它
1. 由于 OpenAI 对大陆 ip 的限制，阁下所用的服务器推荐在大陆以外，或者给服务器套代理
2. 企业微信如果没有做企业备案，那么最多服务100人，这意味着阁下需要「拓展业务」，需要想办法做备案
3. 只针对备案后的企业微信：配置的事件接受服务器，需要和企业微信备案的主体一致。





