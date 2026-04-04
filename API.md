# GopherAI API 文档

## 基础信息

- 基础路径：`/api/v1`
- 数据格式：除图片识别接口外，请求体均为 `application/json`
- 鉴权范围：`/api/v1/AI` 与 `/api/v1/image` 下接口需要 JWT 鉴权
- 鉴权方式：
  - 推荐通过请求头传递：`Authorization: Bearer <token>`
  - 也兼容通过 URL 参数传递：`?token=<token>`

## 统一响应结构

大部分非流式接口都会返回统一 JSON 结构：

```json
{
  "status_code": 1000,
  "status_msg": "success"
}
```

常见字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| status_code | number | 业务状态码 |
| status_msg | string | 状态描述 |

常见状态码：

| 状态码 | 含义 |
| --- | --- |
| 1000 | success |
| 2001 | 请求参数错误 |
| 2006 | 无效的Token |
| 2008 | 验证码错误 |
| 2009 | 记录不存在 |
| 4001 | 服务繁忙 |
| 5001 | 模型不存在 |
| 5002 | 无法打开模型 |
| 5003 | 模型运行失败 |

## 用户相关

### POST `/api/v1/user/register`

接口说明：用户注册，注册成功后直接返回登录 token。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| email | string | 是 | 邮箱 |
| captcha | string | 否 | 邮箱验证码 |
| password | string | 否 | 密码 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "token": "xxx"
}
```

### POST `/api/v1/user/login`

接口说明：用户登录。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| username | string | 否 | 用户名 |
| password | string | 否 | 密码 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "token": "xxx"
}
```

### POST `/api/v1/user/captcha`

接口说明：发送邮箱验证码。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| email | string | 是 | 邮箱 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success"
}
```

## AI 相关接口

以下接口均需要 JWT。

### GET `/api/v1/AI/chat/sessions`

接口说明：获取当前登录用户的会话列表。用户信息从 JWT 中解析，不需要额外传参。

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "sessions": [
    {
      "sessionId": "session-uuid",
      "name": "新的对话",
      "modelType": "deepseek",
      "updateAt": "2026-04-04T10:00:00Z"
    }
  ]
}
```

返回字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| sessionId | string | 会话 ID |
| name | string | 会话标题 |
| modelType | string | 模型类型 |
| updateAt | string | 最近更新时间 |

### POST `/api/v1/AI/chat/send-new-session`

接口说明：创建新会话并发送消息，返回 AI 回复和新会话 ID。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "Information": "AI 回复内容",
  "sessionId": "session-uuid"
}
```

### POST `/api/v1/AI/chat/send`

接口说明：向已有会话发送消息并返回 AI 回复。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| sessionId | string | 是 | 会话 ID |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "Information": "AI 回复内容"
}
```

### POST `/api/v1/AI/chat/history`

接口说明：获取指定会话的聊天历史。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| sessionId | string | 是 | 会话 ID |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "history": [
    {
      "is_user": true,
      "content": "你好"
    },
    {
      "is_user": false,
      "content": "你好，有什么可以帮你？"
    }
  ]
}
```

返回字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| is_user | bool | `true` 表示用户消息，`false` 表示 AI 消息 |
| content | string | 消息内容 |

### POST `/api/v1/AI/chat/send-stream-new-session`

接口说明：创建新会话并以 SSE 方式流式返回回答。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| usingGoogle | bool | 否 | 当前请求体中包含该字段，但流式处理逻辑未实际使用 |
| usingRAG | bool | 否 | 当前请求体中包含该字段，但流式处理逻辑未实际使用 |

响应类型：`text/event-stream`

说明：

- 连接建立后，服务端会先发送一条 `data` 事件，下发新建的 `sessionId`
- 随后继续推送模型生成内容
- 失败时会发送 `error` 事件

首条事件示例：

```text
data: {"sessionId": "session-uuid"}
```

### POST `/api/v1/AI/chat/send-stream`

接口说明：向已有会话发送消息，并以 SSE 方式流式返回回答。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| sessionId | string | 是 | 会话 ID |
| usingGoogle | bool | 否 | 当前请求体中包含该字段，但流式处理逻辑未实际使用 |
| usingRAG | bool | 否 | 当前请求体中包含该字段，但流式处理逻辑未实际使用 |

响应类型：`text/event-stream`

说明：

- 服务端会持续输出流式内容
- 失败时会发送 `error` 事件

### POST `/api/v1/AI/agent/travel_plan`

接口说明：根据旅行需求生成结构化旅游方案。

请求参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| description | string | 是 | 旅行需求描述 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "plan": {
    "mode": "plan",
    "overall_summary": "整体路线、节奏与核心建议概括",
    "flight_price": {
      "summary": "机票价格总结",
      "currency": "CNY",
      "price_range": "1800-2600",
      "booking_tips": [
        "建议提前 2 到 4 周关注价格波动"
      ],
      "raw_text": "原始机票摘要"
    },
    "daily_plans": [
      {
        "day": 1,
        "title": "浅草与东京晴空塔",
        "route": "浅草寺 -> 仲见世商店街 -> 东京晴空塔",
        "transport": "地铁 + 步行",
        "summary": "当天安排概述",
        "attractions": [
          {
            "name": "浅草寺",
            "description": "景点介绍",
            "highlights": [
              "雷门地标",
              "传统参道氛围"
            ],
            "images": [
              {
                "title": "浅草寺正门",
                "url": "https://images.unsplash.com/xxx",
                "source": "Unsplash",
                "source_url": "https://unsplash.com/photos/xxx"
              }
            ]
          }
        ],
        "tips": [
          "建议上午前往"
        ]
      }
    ],
    "sources": [
      "https://unsplash.com/photos/xxx"
    ],
    "notice": "",
    "raw_text": ""
  }
}
```

返回字段说明：

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| mode | string | 返回模式，正常结构化结果为 `plan`，解析失败回退时可能为 `raw` |
| overall_summary | string | 总体概括 |
| flight_price | object | 机票价格与购票建议 |
| daily_plans | array | 每日计划列表 |
| daily_plans[].day | number | 第几天 |
| daily_plans[].title | string | 当天标题 |
| daily_plans[].route | string | 当天路线 |
| daily_plans[].transport | string | 交通方式 |
| daily_plans[].summary | string | 当天摘要 |
| daily_plans[].attractions | array | 当天重点景点 |
| daily_plans[].attractions[].name | string | 景点名称 |
| daily_plans[].attractions[].description | string | 景点介绍 |
| daily_plans[].attractions[].highlights | array | 景点亮点 |
| daily_plans[].attractions[].images | array | 景点图片信息 |
| daily_plans[].attractions[].images[].title | string | 图片标题 |
| daily_plans[].attractions[].images[].url | string | 图片地址 |
| daily_plans[].attractions[].images[].source | string | 图片来源名称 |
| daily_plans[].attractions[].images[].source_url | string | 图片来源链接 |
| daily_plans[].tips | array | 当天提示 |
| sources | array | 来源信息与原始链接 |
| notice | string | 附加提示，正常情况下可为空 |
| raw_text | string | 原始文本兜底内容，正常结构化返回时通常为空 |

## 图片相关接口

以下接口均需要 JWT。

### POST `/api/v1/image/recognize`

接口说明：上传图片并进行识别。

请求参数：`multipart/form-data`

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| image | file | 是 | 图片文件 |

响应示例：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "class_name": "cat"
}
```
