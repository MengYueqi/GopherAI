# GopherAI API 文档

## 基础信息
- 基础路径: `/api/v1`
- 认证: `/AI` 和 `/image` 下接口需要 JWT（通过中间件 `jwt.Auth()`）
- 数据格式: 除图片识别外，均为 `application/json`

## 用户相关
### POST /api/v1/user/register
接口说明: 用户注册
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| email | string | 是 | 邮箱 |
| captcha | string | 否 | 验证码 |
| password | string | 否 | 密码 |

### POST /api/v1/user/login
接口说明: 用户登录
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| username | string | 否 | 用户名 |
| password | string | 否 | 密码 |

### POST /api/v1/user/captcha
接口说明: 发送验证码
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| email | string | 是 | 邮箱 |

## AI 聊天相关 (需 JWT)
### GET /api/v1/AI/chat/sessions
接口说明: 获取当前用户会话列表
请求参数: 无 (从 JWT 中获取 userName)

### POST /api/v1/AI/chat/send-new-session
接口说明: 创建新会话并发送消息
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

### POST /api/v1/AI/chat/send
接口说明: 向已有会话发送消息
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| sessionId | string | 是 | 当前会话 ID |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

### POST /api/v1/AI/chat/history
接口说明: 获取聊天历史
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| sessionId | string | 是 | 当前会话 ID |

### POST /api/v1/AI/chat/send-stream-new-session
接口说明: 创建新会话并以 SSE 方式流式返回回答
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

### POST /api/v1/AI/chat/send-stream
接口说明: 已有会话的 SSE 流式回答
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| question | string | 是 | 用户问题 |
| modelType | string | 是 | 模型类型 |
| sessionId | string | 是 | 当前会话 ID |
| usingGoogle | bool | 否 | 是否使用 Google 搜索 |
| usingRAG | bool | 否 | 是否使用 RAG 检索 |

### POST /api/v1/AI/agent/medical_advice
接口说明: 医疗建议生成
请求参数 (JSON):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| description | string | 是 | 症状描述 |

## 图片识别 (需 JWT)
### POST /api/v1/image/recognize
接口说明: 图片识别
请求参数 (multipart/form-data):
| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| image | file | 是 | 图片文件 |
