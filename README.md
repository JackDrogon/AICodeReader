# AICodeReader

AICodeReader 是一个用 Go 语言开发的代码阅读工具。

## 项目结构

```
AICodeReader/
├── cmd/
│   └── aicodereader/      # 主程序入口
├── pkgs/                  # 核心包
├── scripts/               # 脚本文件
├── docs/                  # 文档目录
└── Makefile              # 构建配置
```

## 快速开始

### 前置要求

- Go 1.19 或更高版本

### 安装

```bash
go mod download
```

### 构建

```bash
make build
```

或者直接使用 Go 命令：

```bash
go build -o bin/aicodereader ./cmd/aicodereader
```

### 运行

```bash
./bin/aicodereader
```

## 开发

### 运行测试

```bash
go test ./...
```

### 代码检查

项目使用 golangci-lint 进行代码质量检查。运行以下命令进行检查：

```bash
golangci-lint run
```

## 贡献

欢迎提交 Pull Request 和 Issue！

## 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。
