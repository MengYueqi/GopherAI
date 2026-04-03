# mcp-photo

`mcp-photo` 是一个基于 `mcp-go` 实现的 MCP 工具服务，用于通过 Unsplash 搜索图片。

它的设计方式与 `mcp-flight` 保持一致：

- 独立 Go 可执行程序
- 支持 `stdio` 和 `sse` 两种传输方式
- 通过环境变量注入第三方 API 凭证
- 返回适合大模型直接消费的文本结果

当前工具名为 `search_photos`。

## 作用

当模型需要“找图片”“找配图”“获取某个主题的参考图”时，可以调用 `search_photos`，由服务向 Unsplash 发起请求并返回结果。

为了减少 token 消耗，当前返回结果只保留每张图片的这几项信息：

- 图片序号
- 描述
- 图片 ID
- 摄影师信息
- `urls.regular`

不再返回 `small`、`thumb`、`download`、`html`、`raw`、`full` 等字段。

## 环境变量

启动前需要设置：

```bash
export UNSPLASH_ACCESS_KEY=your_access_key
```

说明：

- `UNSPLASH_ACCESS_KEY` 为 Unsplash API 的 Access Key
- 服务在启动时不会强制退出，但实际调用 `search_photos` 时如果未设置该变量，会直接报错

## 启动方式

在项目根目录下执行：

```bash
cd common/tools/mcp/eino-mcp/tools/mcp-photo
go build -o mcp-photo main.go
./mcp-photo -transport=sse -server_listen=localhost:8084
```

如果使用标准输入输出模式：

```bash
./mcp-photo -transport=stdio
```

参数说明：

- `-transport`：传输方式，支持 `sse` 或 `stdio`
- `-server_listen`：SSE 模式监听地址，默认 `localhost:8080`

## MCP 工具定义

工具名：

```text
search_photos
```

工具描述：

```text
Search photos via Unsplash
```

## 输入参数

### 必填参数

#### `query`

搜索关键词。

示例：

```json
{
  "query": "snow mountain"
}
```

### 可选参数

#### `page`

页码，默认值为 `1`。

约束：

- 必须是整数
- 必须大于 `0`

#### `per_page`

每页返回数量，默认值为 `10`。

约束：

- 必须是整数
- 当前实现限制为 `1` 到 `30`

#### `order_by`

排序方式。

可选值：

- `relevant`
- `latest`

默认值：

```text
relevant
```

#### `collections`

限制搜索范围到指定收藏集，多个 ID 用英文逗号拼接。

示例：

```text
123,456,789
```

#### `content_filter`

内容安全过滤级别。

可选值：

- `low`
- `high`

默认值：

```text
low
```

#### `color`

颜色过滤。

可选值：

- `black_and_white`
- `black`
- `white`
- `yellow`
- `orange`
- `red`
- `purple`
- `magenta`
- `green`
- `teal`
- `blue`

#### `orientation`

图片方向过滤。

可选值：

- `landscape`
- `portrait`
- `squarish`

## 请求方式

服务内部调用 Unsplash 的接口如下：

```http
GET /search/photos
```

请求地址：

```text
https://api.unsplash.com/search/photos
```

认证方式：

```http
Authorization: Client-ID <UNSPLASH_ACCESS_KEY>
Accept-Version: v1
```

## 示例调用

### 最小调用

```json
{
  "query": "cat"
}
```

### 带筛选条件的调用

```json
{
  "query": "ocean sunset",
  "page": 1,
  "per_page": 5,
  "order_by": "relevant",
  "content_filter": "high",
  "color": "blue",
  "orientation": "landscape"
}
```

## 返回格式

工具返回的是纯文本，不是原始 JSON。

返回结构大致如下：

```text
Found 133 photos across 7 pages for query "coffee".

1. A man drinking a coffee.
ID: eOLpJytrbsQ
Photographer: Jeff Sheldon (@ugmonk)
Regular: https://images.unsplash.com/...

2. ...
```

## 无结果时的返回

如果没有搜索到图片，返回：

```text
no photo results for query "xxx"
```

## 出错行为

常见错误包括：

- `UNSPLASH_ACCESS_KEY is not set`
- `query must be a non-empty string`
- `page must be greater than 0`
- `per_page must be between 1 and 30`
- `order_by must be one of: latest, relevant`
- `content_filter must be one of: low, high`
- `orientation must be one of: landscape, portrait, squarish`
- `unsplash api error: ...`

## 设计说明

本工具没有直接透传 Unsplash 的完整响应，而是做了裁剪，原因如下：

- 减少模型上下文中的无效字段
- 降低 token 消耗
- 保留最关键的可展示图片地址 `regular`
- 让模型在“找图”场景下更容易直接消费结果

如果后续确实需要：

- 原图地址
- 下载链接
- 图片详情页
- 宽高和颜色

可以再按需要增量补回，但默认不建议全部返回。
