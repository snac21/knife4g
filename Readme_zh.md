# Knife4g

Knife4g 是一个用于 Go 语言的 API 文档生成工具，它可以将 OpenAPI 规范转换为美观的 Swagger UI 文档。

## 功能特点

- 支持 OpenAPI 3.0 规范
- 内置美观的 UI 界面
- 支持静态资源嵌入
- 简单易用的配置选项

## 安装

```bash
go get github.com/snac21/knife4g
```

## 使用方法

1. 在您的 HTTP 服务器中集成 Knife4g：

```go
import (
    "github.com/snac21/knife4g"
)

...

// 配置 Knife4g
openAPI := &knife4g.OpenAPI3{}
if content, err := os.ReadFile("./openapi.yaml"); err == nil {
   if err := yaml.Unmarshal(content, openAPI); err != nil {
     stdlog.Printf("Failed to parse OpenAPI document: %v", err)
   }
}

// Configure Knife4g
config := &knife4g.Config{
   RelativePath:   "",           // Access path prefix
   ServerName:    "api-service", // your server name
   OpenAPI:       openAPI,   // OpenAPI document content
}

// Create documentation handler
docHandler := knife4g.Handler(config)

// Register with HTTP server
srv := http.NewServer(opts...)
srv.HandlePrefix("", docHandler)
```

2. 访问文档：
   - 打开浏览器访问 `http://your-server:port/doc.html` 查看 API 文档界面

## 配置说明

- `RelativePath`: 文档访问路径前缀
- `ServerName`: 自定义服务名
- `OpenAPI`: OpenAPI 规范文档内容

## 注意事项

- 确保 OpenAPI 文档格式正确
- 建议在生产环境中配置适当的安全措施
- 静态资源默认使用内置的嵌入文件系统

## 许可证

Apache License

## Acknowledgement
- Thanks to [knife4j](https://github.com/xiaoymin/swagger-bootstrap-ui)
- Thanks to [hononet639](https://github.com/hononet639/knife4g)

