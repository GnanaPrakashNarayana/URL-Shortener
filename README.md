# URL Shortener

A simple URL shortener service written in Go.

## Features

- Shorten long URLs to easily shareable links
- Redirect to original URLs
- Track visit count
- Web interface for shortening URLs
- REST API for programmatic usage

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Installation

1. Clone the repository
   \`\`\`
   git clone https://github.com/yourusername/url-shortener.git
   cd url-shortener
   \`\`\`

2. Build the application
   \`\`\`
   go build -o url-shortener cmd/main.go
   \`\`\`

3. Run the application
   \`\`\`
   ./url-shortener
   \`\`\`

### Configuration

The application can be configured using environment variables or a \`.env\` file:

- \`SERVER_ADDRESS\`: The address on which the server will listen (default: \`:8080\`)
- \`BASE_URL\`: The base URL for shortened links (default: \`http://localhost:8080\`)

## API Documentation

### Shorten a URL

\`\`\`
POST /api/shorten
Content-Type: application/json

{
  "url": "https://example.com/very/long/url"
}
\`\`\`

Response:

\`\`\`
HTTP/1.1 201 Created
Content-Type: application/json

{
  "short_url": "http://localhost:8080/abc123",
  "original_url": "https://example.com/very/long/url",
  "created_at": "2023-01-01T12:00:00Z",
  "visits": 0
}
\`\`\`

### List all URLs

\`\`\`
GET /api/urls
\`\`\`

Response:

\`\`\`
HTTP/1.1 200 OK
Content-Type: application/json

[
  {
    "short_url": "http://localhost:8080/abc123",
    "original_url": "https://example.com/very/long/url",
    "created_at": "2023-01-01T12:00:00Z",
    "visits": 0
  }
]
\`\`\`

### Redirect to original URL

\`\`\`
GET /{id}
\`\`\`

Response:

\`\`\`
HTTP/1.1 302 Found
Location: https://example.com/very/long/url
\`\`\`

## License

This project is licensed under the MIT License - see the LICENSE file for details.