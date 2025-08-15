# 兼容性说明

## 问题描述

当在旧版本 Linux 系统上运行 go-redisbenchmark 时，可能会遇到以下错误：

```bash
./go-redisbenchmark-linux-amd64: /usr/lib64/libc.so.6: version `GLIBC_2.34' not found (required by ./go-redisbenchmark-linux-amd64)
./go-redisbenchmark-linux-amd64: /usr/lib64/libc.so.6: version `GLIBC_2.32' not found (required by ./go-redisbenchmark-linux-amd64)
```

这是因为编译的二进制文件需要较新版本的 GLIBC，而目标系统使用的是较旧版本。

## 解决方案

### 方案 1: 使用兼容性构建脚本 (推荐)

我们提供了专门的构建脚本来生成兼容旧版本系统的二进制文件：

```bash
# 快速构建
chmod +x scripts/quick-build.sh
./scripts/quick-build.sh

# 或者使用完整构建脚本
chmod +x scripts/build-compatible.sh
./scripts/build-compatible.sh
```

### 方案 2: 使用 Docker 构建 (最大兼容性)

如果您的系统有 Docker，可以使用 Docker 构建最大兼容性的版本：

```bash
chmod +x scripts/docker-build.sh
./scripts/docker-build.sh
```

### 方案 3: 手动构建

如果上述方案不可行，可以手动构建：

```bash
# 创建输出目录
mkdir -p dist-compatible

# 方法 1: 静态链接 (推荐)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -extldflags=-static" \
    -o dist-compatible/go-redisbenchmark-linux-amd64-static \
    cmd/redis-benchmark/main.go

# 方法 2: 无 CGO
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o dist-compatible/go-redisbenchmark-linux-amd64-nocgo \
    cmd/redis-benchmark/main.go

# 方法 3: 旧版本兼容
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 go build \
    -ldflags="-s -w" \
    -o dist-compatible/go-redisbenchmark-linux-amd64-old \
    cmd/redis-benchmark/main.go
```

## 构建参数说明

### 关键环境变量

- `CGO_ENABLED=0`: 禁用 CGO，避免链接 C 库
- `GOOS=linux`: 目标操作系统
- `GOARCH=amd64`: 目标架构
- `GOAMD64=v1`: 使用 AMD64 v1 指令集，提高兼容性

### 链接器标志

- `-s`: 去除符号表和调试信息
- `-w`: 去除 DWARF 调试信息
- `-extldflags=-static`: 静态链接所有库

## 推荐使用顺序

1. **静态链接版本** (`*-static`): 无外部依赖，兼容性最好
2. **无 CGO 版本** (`*-nocgo`): 无 C 库依赖，兼容性较好
3. **旧版本兼容版本** (`*-old`): 针对旧版本优化
4. **Docker 构建版本** (`*-docker`): 使用 Go 1.19 构建，兼容性最佳

## 测试兼容性

在目标系统上运行以下命令测试兼容性：

```bash
# 检查是否为动态可执行文件
ldd go-redisbenchmark-linux-amd64-static

# 检查 GLIBC 版本要求
objdump -T go-redisbenchmark-linux-amd64-static | grep GLIBC

# 运行测试
./go-redisbenchmark-linux-amd64-static -h
```

## 支持的系统

以下系统已经过测试或理论上支持：

- **Kylin Linux Advanced Server V10**
- **UOS Server v20**
- **openEuler 20.03 / 22.03 / 24.03**
- **RedHat Enterprise Linux 5.x / 6.x / 7.x / 8.x / 9.x / 10.x**
- **CentOS 5 / 6 / 7 / 8 / 8-Stream / 9-Stream / 10-Stream**
- **Ubuntu 20.04 / 22.04 / 24.04**
- **Debian 10 / 11 / 12 / 13**
- **SUSE Linux Enterprise Server 11 / 12 / 15**
- **openSUSE 11.x / 12.x / 13.x / 42.x / 15.x / tumbleweed**
- **Alpine 3.x / edge**

## 故障排除

### 常见问题

1. **权限问题**
   ```bash
   chmod +x go-redisbenchmark-linux-amd64-static
   ```

2. **架构不匹配**
   ```bash
   # 检查系统架构
   uname -m
   
   # 检查二进制文件架构
   file go-redisbenchmark-linux-amd64-static
   ```

3. **依赖库问题**
   ```bash
   # 检查依赖库
   ldd go-redisbenchmark-linux-amd64-static
   
   # 如果显示 "not a dynamic executable"，说明是静态链接
   ```

### 如果仍有问题

1. **在目标系统上直接编译**
   ```bash
   # 在目标系统上安装 Go 并编译
   go build -o go-redisbenchmark cmd/redis-benchmark/main.go
   ```

2. **使用 Docker 容器运行**
   ```bash
   # 创建包含 Go 的容器
   docker run --rm -it golang:1.19-alpine sh
   
   # 在容器内编译和运行
   ```

3. **检查系统信息**
   ```bash
   # 检查系统版本
   cat /etc/os-release
   
   # 检查 GLIBC 版本
   ldd --version
   
   # 检查内核版本
   uname -r
   ```

## 性能考虑

- **静态链接版本**: 文件较大，但启动快，无依赖问题
- **无 CGO 版本**: 文件较小，兼容性好，推荐用于生产环境
- **Docker 版本**: 兼容性最佳，但文件可能较大

## 维护说明

- 定期更新构建脚本以支持新的系统版本
- 在 CI/CD 中集成兼容性测试
- 收集用户反馈以改进兼容性

## 联系支持

如果遇到兼容性问题，请：

1. 提供目标系统的详细信息
2. 运行 `ldd --version` 和 `uname -a` 的输出
3. 描述具体的错误信息
4. 说明使用的构建方法 