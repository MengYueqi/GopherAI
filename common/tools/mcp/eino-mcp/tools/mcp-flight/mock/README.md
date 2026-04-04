# mcp-flight mock

这是 `mcp-flight` 的本地 mock 版本，用于开发联调，避免频繁消耗真实 SerpAPI 调用次数。

## 功能

- 保留与正式版相同的 MCP tool 名称
- 支持 `current time`
- 支持 `google_flights`
- `google_flights` 返回固定 mock 文本，不会访问外部 API

## 启动

在项目根目录执行：

```bash
go run ./common/tools/mcp/eino-mcp/tools/mcp-flight/mock -transport=sse -server_listen=localhost:8082
```

如果你已经进入目录：

```bash
cd common/tools/mcp/eino-mcp/tools/mcp-flight/mock
go run . -transport=sse -server_listen=localhost:8082
```

启动后地址为：

```text
http://localhost:8082/sse
```

## 项目中的使用方式

你当前项目里 [medicalAgent.go](/Users/mengfanxing/Documents/GopherAI/common/aihelper/medicalAgent.go) 已经把航班 MCP 地址写成：

```go
const flightBaseURL = "http://localhost:8082/sse"
```

所以只要启动这个 mock 服务，现有代码就会自动连到它，不需要再改业务代码。

## Mock 规则

- `SHA -> NRT`：返回两条东京方向的 mock 航班
- `SHA -> KIX`：返回两条大阪方向的 mock 航班
- 其他路线：返回通用 mock 航班结果
- `currency` 未传时默认 `CNY`
- `type` 未传时默认 `2`

## 示例

如果 agent 调用：

```json
{
  "departure_id": "SHA",
  "arrival_id": "NRT",
  "outbound_date": "2026-05-01",
  "currency": "CNY",
  "type": "2"
}
```

就会得到可直接展示的 mock 航班文本结果。
