# Redis 6.0+ 新特性支持

本项目现在支持 Redis 6.0+ 的两个重要新特性：

## 1. ACL (Access Control List) 支持

Redis 6.0+ 引入了基于用户的访问控制，允许为不同用户分配不同的权限。

### 命令行参数

```bash
# 使用ACL用户名和密码
--user <username>          # ACL用户名
--pass <password>          # ACL密码（与传统的-a参数分开）

# 使用Redis URI（推荐）
--uri <redis-uri>      # Redis连接URI，支持ACL和TLS配置
```

### Redis URI 格式

Redis URI是配置连接的推荐方式，支持所有连接参数：

```bash
# 基本ACL认证
redis://username:password@host:port/db

# 启用TLS的ACL认证
redis://username:password@host:port/db?tls=true

# 使用rediss://自动启用TLS
rediss://username:password@host:port/db

# 完整TLS配置
redis://username:password@host:port/db?tls=true&tls-cert=client.crt&tls-key=client.key&tls-ca=ca.crt

# 跳过TLS验证（仅测试环境）
redis://username:password@host:port/db?tls=true&tls-insecure=true
```

### 使用示例

```bash
# 使用Redis URI进行基准测试
./goRedisBenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6379/0" -n 10000 -c 10

# 使用Redis URI连接集群
./goRedisBenchmark --uri "redis://app1:apppass@127.0.0.1:7000/0" -mode cluster -n 10000

# 使用Redis URI连接TLS Redis
./goRedisBenchmark --uri "rediss://benchmark:mypassword@127.0.0.1:6380/0" -n 10000

# 使用Redis URI进行完整配置
./goRedisBenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-cert=certs/client.crt&tls-key=certs/client.key&tls-ca=certs/ca.crt" -n 10000
```

### 传统参数方式

```bash
# 使用ACL用户进行基准测试
./goRedisBenchmark --user benchmark --pass mypassword -n 10000 -c 10

# 使用ACL用户连接集群
./goRedisBenchmark -mode cluster --user app1 --pass apppass -cluster-addrs "127.0.0.1:7000,127.0.0.1:7001"
```

### ACL配置示例

在 `examples/users.acl` 文件中定义了不同的用户角色：

- `admin`: 管理员，拥有所有权限
- `readonly`: 只读用户
- `writer`: 读写用户
- `app1`: 应用专用用户
- `benchmark`: 基准测试专用用户

## 2. TLS (Transport Layer Security) 支持

Redis 6.0+ 支持TLS加密连接，提供传输层安全性。

### 命令行参数

```bash
--tls                    # 启用TLS连接
--tls-cert <file>       # 客户端证书文件路径
--tls-key <file>        # 客户端私钥文件路径
--tls-ca <file>         # CA证书文件路径
--tls-insecure          # 跳过TLS证书验证（仅用于测试）
```

### 使用示例

```bash
# 使用TLS连接到Redis服务器
./goRedisBenchmark --tls --tls-ca certs/ca.crt -n 10000 -c 10

# 使用客户端证书进行双向TLS认证
./goRedisBenchmark --tls --tls-cert certs/client.crt --tls-key certs/client.key --tls-ca certs/ca.crt -n 10000

# 跳过证书验证（仅用于测试环境）
./goRedisBenchmark --tls --tls-insecure -n 10000
```

## 3. 组合使用

ACL和TLS可以同时使用，提供完整的身份验证和传输安全性：

```bash
# 使用TLS + ACL进行安全的基准测试
./goRedisBenchmark \
  --tls \
  --tls-cert certs/client.crt \
  --tls-key certs/client.key \
  --tls-ca certs/ca.crt \
  -u benchmark \
  --acl-pass mypassword \
  -n 100000 \
  -c 50
```

## 4. 设置Redis服务器

### 生成TLS证书

```bash
# 运行证书生成脚本
chmod +x scripts/generate-certs.sh
./scripts/generate-certs.sh
```

### 启动Redis服务器

```bash
# 使用Docker Compose启动支持TLS和ACL的Redis
docker-compose -f docker-compose.redis6.yml up -d
```

### Redis配置文件

参考 `examples/redis6-tls-acl.conf` 配置文件，包含：

- TLS端口配置 (6380)
- 证书文件路径
- ACL文件路径
- 性能调优参数

## 5. 安全注意事项

1. **生产环境**：
   - 不要使用 `--tls-insecure` 参数
   - 使用强密码和适当的ACL权限
   - 定期轮换证书

2. **测试环境**：
   - 可以使用自签名证书
   - 使用 `--tls-insecure` 跳过验证

3. **证书管理**：
   - 私钥文件权限设置为600
   - 证书文件权限设置为644
   - 定期更新证书

## 6. 故障排除

### 常见问题

1. **TLS连接失败**：
   - 检查证书文件路径
   - 验证证书有效性
   - 确认Redis服务器TLS配置

2. **ACL认证失败**：
   - 检查用户名和密码
   - 验证ACL配置文件
   - 确认用户权限设置

3. **性能影响**：
   - TLS会增加一些CPU开销
   - 使用适当的并发数进行测试

### 调试模式

```bash
# 启用详细日志
./goRedisBenchmark --tls --tls-ca certs/ca.crt -n 1000 -c 1 -q
```

## 7. 兼容性

- **Redis版本**: 6.0+
- **Go版本**: 1.23.6+
- **操作系统**: Windows, Linux, macOS

## 8. 示例配置文件

项目包含完整的示例配置：

- `examples/redis6-tls-acl.conf`: Redis服务器配置
- `examples/users.acl`: ACL用户配置
- `docker-compose.redis6.yml`: Docker部署配置
- `scripts/generate-certs.sh`: 证书生成脚本 