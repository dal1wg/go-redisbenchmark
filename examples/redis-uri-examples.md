# Redis URI 使用示例

本文档展示了go-redisbenchmark中Redis URI的各种使用方式。

## URI格式说明

Redis URI遵循标准格式：`scheme://[username:password@]host:port[/db][?query_parameters]`

### 支持的Scheme
- `redis://` - 标准Redis连接
- `rediss://` - 启用TLS的Redis连接

### 支持的查询参数
- `tls=true` - 启用TLS
- `tls-cert=<file>` - 客户端证书文件路径
- `tls-key=<file>` - 客户端私钥文件路径
- `tls-ca=<file>` - CA证书文件路径
- `tls-insecure=true` - 跳过TLS证书验证

## 基本连接示例

### 1. 无认证连接
```bash
# 连接到本地Redis，数据库0
./go-redisbenchmark --uri "redis://127.0.0.1:6379/0" -n 10000

# 连接到本地Redis，数据库1
./go-redisbenchmark --uri "redis://127.0.0.1:6379/1" -n 10000
```

### 2. 传统密码认证
```bash
# 使用传统密码认证
./go-redisbenchmark --uri "redis://:mypassword@127.0.0.1:6379/0" -n 10000

# 等价于
./go-redisbenchmark -a mypassword -n 10000
```

## ACL认证示例

### 3. 基本ACL认证
```bash
# 使用ACL用户名和密码
./go-redisbenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6379/0" -n 10000

# 等价于
./go-redisbenchmark -u benchmark --acl-pass mypassword -n 10000
```

### 4. 不同ACL用户
```bash
# 使用只读用户
./go-redisbenchmark --uri "redis://readonly:readpass@127.0.0.1:6379/0" -n 10000

# 使用应用用户
./go-redisbenchmark --uri "redis://app1:apppass@127.0.0.1:6379/0" -n 10000

# 使用管理员用户
./go-redisbenchmark --uri "redis://admin:adminpass@127.0.0.1:6379/0" -n 10000
```

## TLS连接示例

### 5. 基本TLS连接
```bash
# 使用rediss://自动启用TLS
./go-redisbenchmark --uri "rediss://127.0.0.1:6380/0" -n 10000

# 等价于
./go-redisbenchmark --tls -n 10000
```

### 6. 手动启用TLS
```bash
# 使用查询参数启用TLS
./go-redisbenchmark --uri "redis://127.0.0.1:6380/0?tls=true" -n 10000
```

### 7. 完整TLS配置
```bash
# 指定所有TLS参数
./go-redisbenchmark --uri "redis://127.0.0.1:6380/0?tls=true&tls-cert=certs/client.crt&tls-key=certs/client.key&tls-ca=certs/ca.crt" -n 10000

# 等价于
./go-redisbenchmark --tls --tls-cert certs/client.crt --tls-key certs/client.key --tls-ca certs/ca.crt -n 10000
```

### 8. 跳过TLS验证
```bash
# 跳过TLS证书验证（仅测试环境）
./go-redisbenchmark --uri "redis://127.0.0.1:6380/0?tls=true&tls-insecure=true" -n 10000

# 等价于
./go-redisbenchmark --tls --tls-insecure -n 10000
```

## 组合使用示例

### 9. ACL + TLS
```bash
# ACL认证 + TLS连接
./go-redisbenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true" -n 10000

# 使用rediss:// + ACL
./go-redisbenchmark --uri "rediss://benchmark:mypassword@127.0.0.1:6380/0" -n 10000
```

### 10. 完整配置
```bash
# ACL认证 + 完整TLS配置
./go-redisbenchmark --uri "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-cert=certs/client.crt&tls-key=certs/client.key&tls-ca=certs/ca.crt" -n 10000
```

## 集群模式示例

### 11. 集群连接
```bash
# 使用ACL连接到集群
./go-redisbenchmark --mode cluster --uri "redis://app1:apppass@127.0.0.1:7000/0" -n 10000

# 使用TLS连接到集群
./go-redisbenchmark --mode cluster --uri "rediss://127.0.0.1:7000/0" -n 10000
```

## 哨兵模式示例

### 12. 哨兵连接
```bash
# 使用ACL连接到哨兵
./go-redisbenchmark --mode sentinel --uri "redis://benchmark:mypassword@127.0.0.1:26379/0" --sentinel-addrs "127.0.0.1:26379" --sentinel-master "mymaster" -n 10000
```

## 环境变量支持

### 13. 使用环境变量
```bash
# 设置环境变量
export REDIS_URI="redis://benchmark:mypassword@127.0.0.1:6379/0"

# 使用环境变量
./go-redisbenchmark --uri "$REDIS_URI" -n 10000
```

## 配置文件示例

### 14. 配置文件中的URI
```yaml
# config.yaml
redis:
  uri: "redis://benchmark:mypassword@127.0.0.1:6379/0"
  tests: ["PING", "GET", "SET"]
  requests: 100000
  concurrency: 50
```

## 最佳实践

### 15. 推荐用法
```bash
# 生产环境：使用完整的URI配置
./go-redisbenchmark --uri "redis://appuser:strongpassword@redis.example.com:6379/0?tls=true&tls-cert=/path/to/client.crt&tls-key=/path/to/client.key&tls-ca=/path/to/ca.crt" -n 1000000 -c 100

# 测试环境：使用简化配置
./go-redisbenchmark --uri "redis://benchmark:testpass@127.0.0.1:6379/0" -n 10000 -c 10

# 开发环境：使用基本连接
./go-redisbenchmark --uri "redis://127.0.0.1:6379/0" -n 1000 -c 1
```

## 故障排除

### 常见问题

1. **URI格式错误**
   ```bash
   # 错误：缺少scheme
   ./go-redisbenchmark --uri "127.0.0.1:6379" -n 1000
   
   # 正确：包含scheme
   ./go-redisbenchmark --uri "redis://127.0.0.1:6379" -n 1000
   ```

2. **TLS配置不完整**
   ```bash
   # 错误：启用TLS但未指定证书
   ./go-redisbenchmark --uri "redis://127.0.0.1:6380/0?tls=true" -n 1000
   
   # 正确：指定CA证书或跳过验证
   ./go-redisbenchmark --uri "redis://127.0.0.1:6380/0?tls=true&tls-ca=certs/ca.crt" -n 1000
   ```

3. **ACL认证失败**
   ```bash
   # 检查用户名和密码是否正确
   ./go-redisbenchmark --uri "redis://wronguser:wrongpass@127.0.0.1:6379/0" -n 1000
   ``` 