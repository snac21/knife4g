# Knife4g

Knife4g is an API documentation generation tool for the Go programming language that converts OpenAPI specifications into elegant Swagger UI documentation.

## Features

- Supports OpenAPI 3.0 specification
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

...
// Openapi doc struct
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

2. Access the documentation:
    - Open your browser and visit http://your-server:port/doc.html to view the API documentation interface

## Configuration

- `RelativePath`: Documentation access path prefix
- `ServerName`: Your server name
- `OpenAPI`: OpenAPI specification document content

## Notes

- Ensure OpenAPI document format is correct
- Recommended to configure appropriate security measures in production environments
- Static resources use built-in embedded filesystem by default

## License

Apache License

## Acknowledgement
- Thanks to [knife4j](https://github.com/xiaoymin/swagger-bootstrap-ui)
- Thanks to [hononet639](https://github.com/hononet639/knife4g)