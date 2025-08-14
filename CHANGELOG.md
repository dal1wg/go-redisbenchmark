# 变更日志

## [2025-01-XX] - 修复版本

### 修复的问题

#### 1. 修复 .air.toml 配置文件
- 删除了与项目不相关的旧配置文件
- 创建了新的 .air.toml 配置文件，专门为 go-redisbenchmark 项目配置
- 配置了正确的构建命令和输出路径
- 添加了适当的文件类型过滤（.go, .mod, .sum）

#### 2. 为压测过程中的 Redis key 设置淘汰时间
- 为所有预填充的数据类型（String、Hash、List、Set、ZSet、Bitmap、HyperLogLog、Geo、Stream）设置了可配置的 TTL
- 默认 TTL 为 1 小时，可通过 `--ttl` 参数自定义
- 确保压测完成后，所有测试数据会在指定时间内自动过期
- 防止压测过程中产生的历史数据在 Redis 中残留
- 使用 `pipe.Expire()` 命令为复合数据类型设置过期时间
- 为 SET 测试操作也设置了相同的过期时间

### 技术细节

- TTL 默认为 1 小时（3600 秒），可通过命令行参数自定义
- 支持标准时间格式：1h, 30m, 2h, 3600s 等
- 使用 Redis Pipeline 批量设置过期时间，提高性能
- 保持了原有的压测逻辑不变，只是添加了数据生命周期管理
- 所有 key 都使用 `{gorb}:` 前缀，便于识别和管理

### 影响

- **正面影响**：
  - 自动清理压测数据，无需手动清理
  - 减少 Redis 内存占用
  - 提高压测环境的可重复性
  
- **注意事项**：
  - 压测数据会在 1 小时后自动过期
  - 如果需要更长的数据保留时间，可以调整 TTL 值
  - 建议在压测完成后等待 1 小时，或手动清理数据

### 使用方法

1. 使用 air 进行热重载开发：
   ```bash
   air
   ```

2. 运行压测（数据会自动设置 1 小时过期时间）：
   ```bash
   go run cmd/redis-benchmark/main.go -n 10000 -c 10
   ```

3. 自定义 TTL 时间：
   ```bash
   # 设置 30 分钟过期
   go run cmd/redis-benchmark/main.go -n 10000 -c 10 --ttl 30m
   
   # 设置 2 小时过期
   go run cmd/redis-benchmark/main.go -n 10000 -c 10 --ttl 2h
   
   # 设置 1800 秒过期
   go run cmd/redis-benchmark/main.go -n 10000 -c 10 --ttl 1800s
   ```

4. 压测完成后，数据会在指定时间内自动过期，无需手动清理 