# Rudimentary HTTP/1.1 in Go

This repository contains a very basic, from-scratch implementation of HTTP/1.1 written in Go. It was created purely for educational purposes, to better understand how HTTP works under the hood and to experiment with sockets and request parsing.

- Very basic handling of HTTP/1.1 requests
- Minimal parsing of HTTP headers
- Persistent connections (keep-alive)
- Raw TCP socket usage
- No external libraries

## Usage

Clone the repo and run:

```bash
go run main.go
```

Then send a request:
```bash
```bash
curl -v http://localhost:4221/echo/test
```

