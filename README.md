# Crawler

A simple web crawler written in Go.

## Features

- Fetches and parses web pages
- Extracts links and resources
- Supports recursive crawling
- Configurable depth and concurrency

## Getting Started

### Prerequisites

- Go 1.18 or higher

### Installation

```bash
git clone https://github.com/yourusername/crawler.git
cd crawler
go build
```

### Usage

```bash
./crawler -url https://example.com -depth 2
```

#### Command-line Flags

- `-url` : Starting URL to crawl (required)
- `-depth` : Maximum crawl depth (default: 1)
- `-concurrency` : Number of concurrent workers (default: 5)

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

This project is licensed under the MIT License.