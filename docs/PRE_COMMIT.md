# Pre-commit Hook 配置

本项目使用 [pre-commit](https://pre-commit.com/) 工具来确保代码质量，防止不符合 CI 标准的代码被提交。

## 安装和设置

### 1. 安装 pre-commit

```bash
pip install pre-commit
```

### 2. 安装 git hooks

```bash
pre-commit install
```

### 3. 运行检查（可选）

可以手动运行所有检查：

```bash
pre-commit run --all-files
```

## 工作原理

pre-commit hook 配置在 `.pre-commit-config.yaml` 文件中，包含以下检查：

1. **基本文件检查**：
   - 移除尾随空白符
   - 确保文件以换行符结尾
   - 检查 YAML 语法
   - 检查大文件
   - 检查合并冲突标记

2. **CI 流水线检查**：
   - 通过 `make ci` 运行完整的 CI 检查
   - 包括：代码格式检查、linting、测试、构建

## 提交行为

当您尝试提交代码时：

- ✅ 如果所有检查通过，提交成功
- ❌ 如果任何检查失败，提交被阻止
- 🔧 某些检查（如格式化）会自动修复问题

## 示例

```bash
$ git commit -m "fix: update function"
# 如果代码格式有问题或测试失败，提交会被阻止
# 修复问题后重新提交即可
```

## 跳过 pre-commit（不推荐）

如果在特殊情况下需要跳过 pre-commit 检查：

```bash
git commit --no-verify -m "emergency fix"
```

**注意**：只在紧急情况下使用，并确保后续及时修复问题。
