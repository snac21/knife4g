# Knife4g

Knife4g is an API documentation generation tool for the Go programming language that converts OpenAPI specifications into elegant Swagger UI documentation.

## Features

- Supports OpenAPI 3.0 specification
- Automatic conversion to Swagger 2.0 format
- Built-in elegant UI interface
- Supports static resource embedding
- Simple and easy-to-use configuration options

## Installation

```bash
go get github.com/snac21/knife4g
```

## Usage

1. Integrate Knife4g in your HTTP server:

```go
import (
    "github.com/snac21/knife4g"
)

var (
	OpenApiContent []byte
)

...
OpenApiContent, _ = os.ReadFile(yourOpenApiPath)

// Configure Knife4g
config := &knife4g.Config{
    RelativePath:   "/doc",           // Access path prefix
    OpenApiContent: openApiContent,   // OpenAPI document content
    KService: knife4g.Knife4gService{
        Name:           "API Documentation",
        Url:            "/docYaml",
        Location:       "/docYaml",
        SwaggerVersion: "2.0",
    },
}

// Create documentation handler
docHandler := knife4g.Handler(config)

// Register with HTTP server
srv := http.NewServer(opts...)
srv.HandlePrefix("/doc", docHandler)
```

2. Access the documentation:
    - Open your browser and visit http://your-server:port/doc/index to view the API documentation interface

## Configuration

- `RelativePath`: Documentation access path prefix
- `OpenApiContent`: OpenAPI specification document content
- `KService`: Service configuration information

## Notes

- Ensure OpenAPI document format is correct
- Recommended to configure appropriate security measures in production environments
- Static resources use built-in embedded filesystem by default

## License

Apache License

## Acknowledgement
Thanks to [knife4j](https://github.com/xiaoymin/swagger-bootstrap-ui)
Thanks to [hononet639](https://github.com/hononet639/knife4g)