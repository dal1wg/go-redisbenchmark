# Go Redis Benchmark

一个高性能的Redis基准测试工具，用Go语言编写，支持Redis 6.0+的新特性。

## 特性

- 🚀 **高性能**: 支持高并发和管道操作
- 🔐 **Redis 6.0+ ACL支持**: 基于用户的访问控制
- 🔒 **TLS加密**: 支持传输层安全连接
- 🌐 **Redis URI支持**: 标准化的连接字符串配置
- 🏗️ **多模式支持**: 单机、集群、哨兵模式
- 📊 **全面测试**: 支持所有Redis数据类型的基准测试
- 🐳 **Docker支持**: 提供完整的容器化部署方案

## Redis 6.0+ 新特性支持

### ACL (Access Control List)
- 支持用户名/密码认证
- 与传统的密码认证兼容
- 命令行参数: `--user <username> --pass <password>`
- **Redis URI支持**: `--uri "redis://username:password@host:port/db"`

### TLS (Transport Layer Security)
- 支持加密连接
- 客户端证书认证
- CA证书验证
- 命令行参数: `--tls --tls-cert --tls-key --tls-ca`
- **Redis URI支持**: `--uri "rediss://host:port/db"` 或 `--uri "redis://host:port/db?tls=true"`

### Redis URI 格式
支持标准的Redis URI格式，简化配置：
```bash
# 基本ACL认证
redis://username:password@host:port/db

# TLS连接
rediss://host:port/db

# 完整配置
redis://username:password@host:port/db?tls=true&tls-cert=client.crt&tls-key=client.key&tls-ca=ca.crt
```

详细使用说明请参考 [Redis 6.0+ 特性文档](docs/redis6-features.md) 和 [Redis URI 示例](examples/redis-uri-examples.md)

## 快速开始

支持 Centos/Ubuntu/Debain/Windows/Macos

### 构建

运行构建和测试:
```bash
make all
```

构建应用:
```bash
make build
```

跨平台构建 (输出到 dist/):
```bash
make build-all
```

运行测试套件:
```bash
make test
```

清理构建产物:
```bash
make clean
```

### 运行

#### 基本用法
- 通过Go直接运行 (跨平台):
```bash
go run cmd/redis-benchmark/main.go --read-all -n 100000 -c 200 -P 50
```

- 运行构建后的二进制文件 (Windows):
```powershell
./go-redisbenchmark.exe --read-all -n 100000 -c 200 -P 50
```

- 运行构建后的二进制文件 (Linux/macOS):
```bash
./go-redisbenchmark --read-all -n 100000 -c 200 -P 50
```

#### 使用ACL和TLS
```bash
# 使用ACL用户
./go-redisbenchmark --user benchmark --pass mypassword -n 10000 -c 10

# 使用TLS连接
./go-redisbenchmark --tls --tls-ca certs/ca.crt -n 10000

# 组合使用ACL和TLS
./go-redisbenchmark \
  --tls \
  --tls-cert certs/client.crt \
  --tls-key certs/client.key \
  --tls-ca certs/ca.crt \
  --user benchmark \
  --pass mypassword \
  -n 100000 \
  -c 50

# 使用Redis URI（推荐方式）
./go-redisbenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6379/0" -n 10000

# 使用Redis URI + TLS
./go-redisbenchmark --uri "rediss://benchmark:mypassword@127.0.0.1:6380/0" -n 10000

# 使用Redis URI完整配置
./go-redisbenchmark \
  --uri "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-cert=certs/client.crt&tls-key=certs/client.key&tls-ca=certs/ca.crt" \
  -n 100000 \
  -c 50
```

### 开发模式

安装air后或通过Makefile助手:
```bash
make watch
```
这将在文件变化时运行 `go run cmd/redis-benchmark/main.go`。

## 配置示例

项目包含完整的配置示例:

- `examples/redis6-tls-acl.conf`: Redis服务器配置
- `examples/users.acl`: ACL用户配置  
- `docker-compose.redis6.yml`: Docker部署配置
- `scripts/generate-certs.sh`: TLS证书生成脚本

## 许可证

BSD 3-Clause License

Copyright (c) 2025, dalew

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

1. Redistributions of source code must retain the above copyright notice, this
   list of conditions and the following disclaimer.

2. Redistributions in binary form must reproduce the above copyright notice,
   this list of conditions and the following disclaimer in the documentation
   and/or other materials provided with the distribution.

3. Neither the name of the copyright holder nor the names of its
   contributors may be used to endorse or promote products derived from
   this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
