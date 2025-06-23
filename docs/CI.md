# CI/CD 设置指南

这个项目已经配置了完整的CI/CD流程，包括代码质量检查、测试和构建。使用golangci-lint v2.1.6作为代码质量检查工具。

## 自动化 CI/CD (GitHub Actions)

### 触发条件
- 推送到 `master` 或 `main` 分支
- 创建 Pull Request 到 `master` 或 `main` 分支

### CI 流程包括以下步骤：

#### 1. 测试 (Test Job)
- 下载Go依赖
- 运行单元测试（包含竞态检测）
- 生成覆盖率报告
- 上传覆盖率到Codecov

#### 2. 代码检查 (Lint Job)  
- 使用 `golangci-lint` 进行代码质量检查
- 配置文件：`.golangci.yml`

#### 3. 构建 (Build Job)
- 编译二进制文件
- 上传构建产物

#### 4. 格式检查 (Format Check Job)
- 检查代码格式是否符合Go标准

## 本地开发

### 快速开始
```bash
# 运行所有CI检查
make ci

# 单独运行各个步骤
make check-fmt    # 检查代码格式
make lint         # 代码质量检查
make test         # 运行测试
make build        # 构建项目
```

### 代码格式化
```bash
# 格式化代码 (使用go fmt)
make fmt

# 检查格式
make check-fmt

# 使用golangci-lint v2的新格式化功能
golangci-lint fmt
```

### 测试
```bash
# 运行测试
make test

# 生成覆盖率报告
make test-coverage
```

### 代码质量检查
```bash
# 运行linter
make lint
```

## 工具要求

### 必需工具
- Go 1.24.x
- golangci-lint v2.1.6+

### 安装 golangci-lint
```bash
# 使用脚本安装
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

# 或者使用go install (推荐)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 验证安装
golangci-lint version
```

## 配置文件

- `.github/workflows/ci.yml` - GitHub Actions CI配置
- `.golangci.yml` - golangci-lint v2配置文件
- `Makefile` - 本地开发命令
- `scripts/check-dev-env.sh` - 开发环境检查脚本

## 代码质量标准

### golangci-lint v2新特性：
- **新的配置格式**: 更清晰、更易维护的配置结构
- **格式化器**: 新增`golangci-lint fmt`命令支持代码格式化
- **改进的文件路径管理**: 路径相对于配置文件而非执行目录
- **预设排除规则**: 人性化的排除规则预设
- **更好的性能**: 优化的linter执行性能

### 当前启用的Linters包括：
- **基础检查**: errcheck, govet, staticcheck, ineffassign, unused
- **代码风格**: revive, whitespace, unconvert
- **安全检查**: gosec
- **性能优化**: prealloc
- **复杂度控制**: gocyclo, gocognit, funlen, nestif
- **错误处理**: errorlint, errname
- **现代Go特性**: copyloopvar, intrange, mirror, sloglint
- **最佳实践**: misspell, godot, asciicheck, bidichk

### 排除规则
- 测试文件的某些严格检查被放松
- Vendor目录被完全排除
- 某些常见的非关键问题被忽略

## golangci-lint v2配置迁移

如果你有旧的golangci-lint v1配置文件，可以使用迁移命令：

```bash
# 自动迁移配置文件到v2格式
golangci-lint migrate

# 跳过验证进行迁移
golangci-lint migrate --skip-validation
```

## 开发环境检查

运行环境检查脚本来验证你的开发环境：

```bash
# 检查开发环境
make check-env
```

## 提交代码前的检查清单

1. ✅ 运行 `make check-env` 检查开发环境
2. ✅ 运行 `make fmt` 格式化代码
3. ✅ 运行 `make lint` 检查代码质量  
4. ✅ 运行 `make test` 确保所有测试通过
5. ✅ 运行 `make ci` 执行完整检查

遵循这些步骤将确保你的代码能够通过CI检查。 