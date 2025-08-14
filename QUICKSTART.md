# 快速开始指南 - Redis 6.0+ ACL和TLS

本指南将帮助您快速设置和测试Redis 6.0+的ACL和TLS功能。

## 前置要求

- Docker 和 Docker Compose
- Go 1.23.6+
- OpenSSL (用于生成证书)

## 步骤1: 生成TLS证书

```bash
# 生成测试用的TLS证书
make certs
```

这将创建以下文件：
- `certs/ca.crt` - CA证书
- `certs/redis.crt` - Redis服务器证书
- `certs/redis.key` - Redis服务器私钥
- `certs/client.crt` - 客户端证书
- `certs/client.key` - 客户端私钥

## 步骤2: 启动Redis 6.0+服务器

```bash
# 启动支持TLS和ACL的Redis服务器
make redis6-up
```

这将启动：
- Redis服务器在端口6379 (非TLS)
- Redis服务器在端口6380 (TLS)
- 预配置的ACL用户

## 步骤3: 测试功能

### 测试基本连接
```bash
# 测试非TLS连接
go run cmd/redis-benchmark/main.go -n 1000 -c 1
```

### 测试ACL认证
```bash
# 使用ACL用户进行基准测试
go run cmd/redis-benchmark/main.go -u benchmark --acl-pass mypassword -n 1000 -c 1

# 使用Redis URI进行ACL认证（推荐）
go run cmd/redis-benchmark/main.go --uri "redis://benchmark:mypassword@127.0.0.1:6379/0" -n 1000 -c 1
```

### 测试TLS连接
```bash
# 使用TLS连接进行基准测试
go run cmd/redis-benchmark/main.go --tls --tls-ca certs/ca.crt -n 1000 -c 1

# 使用Redis URI进行TLS连接
go run cmd/redis-benchmark/main.go --uri "rediss://127.0.0.1:6380/0" -n 1000 -c 1
```

### 测试组合功能
```bash
# 同时使用ACL和TLS
go run cmd/redis-benchmark/main.go \
  --tls \
  --tls-cert certs/client.crt \
  --tls-key certs/client.key \
  --tls-ca certs/ca.crt \
  -u benchmark \
  --acl-pass mypassword \
  -n 1000 \
  -c 1

# 使用Redis URI进行完整配置（推荐）
go run cmd/redis-benchmark/main.go \
  --uri "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-cert=certs/client.crt&tls-key=certs/client.key&tls-ca=certs/ca.crt" \
  -n 1000 \
  -c 1
```

## 步骤4: 运行完整测试套件

```bash
# 运行Redis 6.0+功能测试
make test-redis6
```

## 步骤5: 清理

```bash
# 停止Redis容器
make redis6-down

# 清理构建产物
make clean
```

## 故障排除

### 证书问题
如果遇到TLS证书错误：
```bash
# 重新生成证书
make certs

# 检查证书文件权限
ls -la certs/
```

### 连接问题
如果无法连接到Redis：
```bash
# 检查容器状态
docker ps

# 查看容器日志
docker logs redis6-tls-acl
```

### ACL认证失败
如果ACL认证失败：
```bash
# 检查ACL配置
docker exec redis6-tls-acl cat /usr/local/etc/redis/users.acl

# 测试ACL用户
docker exec -it redis6-tls-acl redis-cli -u redis://benchmark:mypassword@127.0.0.1:6379
```

## 生产环境注意事项

1. **不要使用自签名证书**
2. **使用强密码**
3. **定期轮换证书**
4. **限制ACL用户权限**
5. **监控连接日志**

## 下一步

- 阅读 [Redis 6.0+ 特性文档](docs/redis6-features.md)
- 查看 [完整配置示例](examples/)
- 了解 [命令行参数](cmd/redis-benchmark/main.go)

## 获取帮助

如果遇到问题，请检查：
1. Docker容器状态
2. 证书文件是否存在
3. 端口是否被占用
4. ACL配置文件格式 