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
git clone https://github.com/HarshAg90/GoPoweredCrawler.git
cd crawler
go build
```

### Usage

```bash
./crawler -url https://example.com -depth 2
```
> watch video at [video](https://github.com/HarshAg90/GoPoweredCrawler/blob/main/examples/Crawler%20Program%20VID.mp4).

#### Command-line Flags

- `-url` : Starting URL to crawl (required)
- `-depth` : Maximum crawl depth (default: 1)
- `-concurrency` : Number of concurrent workers (default: 5)

## Contributing

Contributions are welcome! Please open issues or submit pull requests.

## License

This project is licensed under the MIT License.