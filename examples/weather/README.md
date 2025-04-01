# 彩云天气 MCP 示例

这个示例展示了如何使用go-mcp库创建一个调用彩云天气API的MCP服务。

## 功能

- 提供`getWeather`工具，用于获取指定位置的天气信息
- 支持实时天气、空气质量、降水、风速等数据
- 支持天气预警信息
- 支持未来小时预报

## 前提条件

你需要一个彩云天气API密钥。可以在[彩云天气开发者中心](https://dashboard.caiyunapp.com/)申请。

## 使用方法

### 启动服务器

```bash
# 设置API密钥
export CAIYUN_API_KEY=你的彩云天气API密钥

# 启动服务器
go run main.go
```

服务器将在以下地址启动：
- HTTP服务器：http://localhost:8080
- WebSocket服务器：ws://localhost:8081

### 使用客户端

```bash
# 使用默认坐标（北京）
go run client/main.go

# 使用自定义坐标（经度 纬度）
go run client/main.go 116.407526 39.90403
```

## API参数

调用`getWeather`工具时需要提供以下参数：

- `longitude`：经度（必填）
- `latitude`：纬度（必填）
- `language`：语言，可选，默认为`zh_CN`，支持`en_US`、`ja`等

## 返回数据

返回的数据包括：

- 实时天气信息（温度、湿度、天气状况等）
- 空气质量信息
- 降水信息
- 天气预警信息（如有）
- 未来24小时天气预报
- 未来天气预报

具体字段说明请参考[彩云天气API文档](https://docs.caiyunapp.com/docs/api)。

## 注意事项

- 使用前请确保已设置环境变量`CAIYUN_API_KEY`
- 彩云天气API有调用频率限制，请合理使用