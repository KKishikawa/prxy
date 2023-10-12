# Prxy - Simple Proxy Server

## Description

This is a simple proxy server that can be used to forward requests to a different server. It is written in go.

## Usage

Set options via command line flags or omit them to input with prompts.

### options

| flag      | description                        |
| --------- | ---------------------------------- |
| -h, -help | help                               |
| -p        | port number                        |
| -t        | target host to forward requests to |

Port number and Target host are required, but can be set via prompts if not set via flags.

## build

```bash
go build main.go
```