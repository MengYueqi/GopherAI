# GopherAI Mock API 文档

这个文件用于前端联调和页面能力测试，不代表后端真实返回，只提供一组稳定、可复现的 mock 数据。

## 启动方式

在项目根目录执行：

```bash
go run ./test/mock
```

默认启动地址：

```text
http://127.0.0.1:9090
```

如需修改端口：

```bash
MOCK_PORT=9001 go run ./test/mock
```

## 基础约定

- 基础路径：`/api/v1`
- 认证方式：需要鉴权的接口统一带 `Authorization: Bearer mock-jwt-token`
- 非流式接口返回 JSON
- 流式接口返回 `text/event-stream`
- 所有成功响应统一使用：

```json
{
  "status_code": 1000,
  "status_msg": "success"
}
```

## 通用 Mock 数据

- Token：`mock-jwt-token`
- 用户名：`mock_user`
- 默认模型：`deepseek-chat`
- 会话 ID 1：`session_mock_001`
- 会话 ID 2：`session_mock_002`

## 用户相关

### POST `/api/v1/user/register`

请求：

```json
{
  "email": "test@example.com",
  "captcha": "123456",
  "password": "mock_password"
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "token": "mock-jwt-token"
}
```

失败响应示例：

```json
{
  "status_code": 2008,
  "status_msg": "验证码错误"
}
```

### POST `/api/v1/user/login`

请求：

```json
{
  "username": "mock_user",
  "password": "mock_password"
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "token": "mock-jwt-token"
}
```

失败响应示例：

```json
{
  "status_code": 2004,
  "status_msg": "用户名或密码错误"
}
```

### POST `/api/v1/user/captcha`

请求：

```json
{
  "email": "test@example.com"
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success"
}
```

## AI 会话相关

以下接口都需要请求头：

```http
Authorization: Bearer mock-jwt-token
```

### GET `/api/v1/AI/chat/sessions`

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "sessions": [
    {
      "sessionId": "session_mock_001",
      "name": "东京三日游攻略",
      "modelType": "deepseek-chat",
      "updateAt": "2026-04-04T09:30:00Z"
    },
    {
      "sessionId": "session_mock_002",
      "name": "感冒用药建议",
      "modelType": "deepseek-chat",
      "updateAt": "2026-04-04T08:10:00Z"
    }
  ]
}
```

空列表响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "sessions": []
}
```

### POST `/api/v1/AI/chat/send-new-session`

请求：

```json
{
  "question": "帮我规划一次东京三日游",
  "modelType": "deepseek-chat",
  "usingGoogle": true,
  "usingRAG": false
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "Information": "当然可以。第一天建议游览浅草寺和晴空塔，第二天前往涩谷、原宿，第三天安排上野公园和秋叶原。",
  "sessionId": "session_mock_003"
}
```

失败响应示例：

```json
{
  "status_code": 5003,
  "status_msg": "模型运行失败"
}
```

### POST `/api/v1/AI/chat/send`

请求：

```json
{
  "question": "预算控制在 5000 元以内",
  "modelType": "deepseek-chat",
  "sessionId": "session_mock_001",
  "usingGoogle": false,
  "usingRAG": false
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "Information": "如果预算控制在 5000 元以内，建议选择商务酒店，优先购买地铁通票，并减少高价景点和高端餐饮安排。"
}
```

会话不存在响应：

```json
{
  "status_code": 2009,
  "status_msg": "记录不存在"
}
```

### POST `/api/v1/AI/chat/history`

请求：

```json
{
  "sessionId": "session_mock_001"
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "history": [
    {
      "is_user": true,
      "content": "帮我规划一次东京三日游"
    },
    {
      "is_user": false,
      "content": "当然可以。第一天建议游览浅草寺和晴空塔，第二天前往涩谷、原宿，第三天安排上野公园和秋叶原。"
    },
    {
      "is_user": true,
      "content": "预算控制在 5000 元以内"
    },
    {
      "is_user": false,
      "content": "如果预算控制在 5000 元以内，建议选择商务酒店，优先购买地铁通票，并减少高价景点和高端餐饮安排。"
    }
  ]
}
```

### POST `/api/v1/AI/agent/travel_plan`

请求：

```json
{
  "description": "从上海出发，东京 3 日游，想看经典景点，预算适中"
}
```

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "plan": {
    "mode": "plan",
    "overall_summary": "东京 3 日行程以经典城市地标、商业街区和文化体验为主，节奏中等，适合第一次到东京旅行的用户。",
    "flight_price": {
      "summary": "往返东京机票通常在淡季更划算，建议优先关注直飞与中转时长之间的平衡。",
      "currency": "CNY",
      "price_range": "1800-2600",
      "booking_tips": [
        "建议提前 2 到 4 周关注价格波动",
        "若预算敏感，可优先考虑非黄金时段航班"
      ],
      "raw_text": "Mock 航班价格区间：1800-2600 CNY，直飞更省时，中转更省预算。"
    },
    "daily_plans": [
      {
        "day": 1,
        "title": "浅草与东京晴空塔",
        "route": "浅草寺 -> 仲见世商店街 -> 隅田公园 -> 东京晴空塔",
        "transport": "地铁 + 步行",
        "summary": "第一天适合从东京传统街区开始，感受寺庙文化与城市天际线。",
        "attractions": [
          {
            "name": "浅草寺",
            "description": "东京代表性的历史寺庙，适合体验传统建筑、参拜文化与街区氛围。",
            "highlights": [
              "雷门地标",
              "传统参道氛围",
              "适合拍照与体验和风街景"
            ],
            "images": [
              {
                "title": "浅草寺正门",
                "url": "https://images.unsplash.com/photo-1542051841857-5f90071e7989",
                "source": "Unsplash",
                "source_url": "https://unsplash.com/photos/OQMZwNd3ThU"
              }
            ]
          }
        ],
        "tips": [
          "浅草区域建议上午前往，人流更可控"
        ]
      }
    ],
    "sources": [
      "https://unsplash.com/photos/OQMZwNd3ThU",
      "https://unsplash.com/photos/JmuyB_LibRo"
    ],
    "notice": "这是 mock 返回的结构化旅游方案，用于前端联调。",
    "raw_text": ""
  }
}
```

结构说明：

- `plan.mode = "plan"` 表示结构化旅游方案
- `plan.overall_summary` 用于顶部总览
- `plan.flight_price` 用于机票价格卡片
- `plan.daily_plans` 用于每日行程渲染
- `plan.daily_plans[].attractions[].images` 用于图片展示
- `plan.sources` 用于来源信息展示
- 如果后端回退失败兜底，可返回 `mode: "raw"` 和 `raw_text`

## 流式接口 Mock

前端可直接按 `EventSource` 或 fetch + stream 的方式消费以下数据。

### POST `/api/v1/AI/chat/send-stream-new-session`

请求：

```json
{
  "question": "帮我规划一次东京三日游",
  "modelType": "deepseek-chat",
  "usingGoogle": true,
  "usingRAG": false
}
```

Mock SSE 输出：

```text
data: {"sessionId":"session_mock_003"}

data: {"content":"当然可以，"}

data: {"content":"第一天建议游览浅草寺和晴空塔，"}

data: {"content":"第二天前往涩谷和原宿，"}

data: {"content":"第三天安排上野公园和秋叶原。"}

data: [DONE]
```

### POST `/api/v1/AI/chat/send-stream`

请求：

```json
{
  "question": "预算控制在 5000 元以内",
  "modelType": "deepseek-chat",
  "sessionId": "session_mock_001",
  "usingGoogle": false,
  "usingRAG": false
}
```

Mock SSE 输出：

```text
data: {"content":"如果预算控制在 5000 元以内，"}

data: {"content":"建议选择商务酒店，"}

data: {"content":"优先购买地铁通票，"}

data: {"content":"并减少高价景点和高端餐饮安排。"}

data: [DONE]
```

流式失败示例：

```text
event: error
data: {"message":"Failed to send message"}
```

## 图片识别

### POST `/api/v1/image/recognize`

请求类型：`multipart/form-data`

字段：

| 字段 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| image | file | 是 | 上传图片 |

成功响应：

```json
{
  "status_code": 1000,
  "status_msg": "success",
  "class_name": "golden_retriever"
}
```

参数错误响应：

```json
{
  "status_code": 2001,
  "status_msg": "请求参数错误"
}
```

## 前端建议覆盖场景

- 登录成功和登录失败
- 注册成功和验证码错误
- 会话列表为空与非空
- 新建会话成功
- 已有会话继续追问
- 聊天历史回显
- 结构化旅游方案渲染
- 每日景点图片展示
- 流式输出逐段渲染
- 流式失败提示
- 图片上传成功和参数缺失
- token 失效时统一跳转登录

## Token 失效 Mock

任意鉴权接口都可以模拟返回：

```json
{
  "status_code": 2006,
  "status_msg": "无效的Token"
}
```
