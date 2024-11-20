# Custom Transfer Protocol

A simple protocol for transferring files or messages between devices using a server.

## 🚀 Getting Started

### Server

```bash
go run ./server
```

### Client

#### Send a message

```bash
go run ./client --send "Hello, world!" --channel 1
```

#### Send a file

```bash
go run ./client --send /path/to/file --channel 1
```

#### Receive files

```bash
go run ./client --receive --channel 1
```

## 📦 Stack

- [Go](https://go.dev/)

## 🙌 Contributors

- [mauroquinteros](https://github.com/mauroquinteros)
