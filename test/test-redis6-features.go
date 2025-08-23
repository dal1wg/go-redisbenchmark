/*
 * go-redisbenchmark - A high performance Redis benchmark tool written in Go.
 * Supports Redis 6.0+ features including ACL and TLS.
 *
 * Copyright (c) 2025, dal1wg <15072697283@139.com>
 * All rights reserved.
 *
 * This source code is licensed under the BSD 3-Clause License.
 */

 package main

 import (
	 "context"
	 "crypto/tls"
	 "crypto/x509"
	 "fmt"
	 "log"
	 "os"
	 "time"
 
	 "github.com/redis/go-redis/v9"
 )
 
 func main() {
	 fmt.Println("测试Redis 6.0+ ACL和TLS功能...")
 
	 // 测试1: 基本连接
	 testBasicConnection()
 
	 // 测试2: ACL认证
	 testACLAuthentication()
 
	 // 测试3: TLS连接
	 testTLSConnection()
 
	 // 测试4: 组合使用ACL和TLS
	 testACLAndTLS()
 }
 
 func testBasicConnection() {
	 fmt.Println("\n=== 测试1: 基本连接 ===")
	 
	 client := redis.NewClient(&redis.Options{
		 Addr:     "127.0.0.1:6379",
		 Password: "",
		 DB:       0,
	 })
 
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()
 
	 _, err := client.Ping(ctx).Result()
	 if err != nil {
		 fmt.Printf("❌ 基本连接失败: %v\n", err)
		 return
	 }
	 fmt.Println("✅ 基本连接成功")
 
	 client.Close()
 }
 
 func testACLAuthentication() {
	 fmt.Println("\n=== 测试2: ACL认证 ===")
	 
	 client := redis.NewClient(&redis.Options{
		 Addr:     "127.0.0.1:6379",
		 Username: "benchmark",
		 Password: "mypassword",
		 DB:       0,
	 })
 
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()
 
	 _, err := client.Ping(ctx).Result()
	 if err != nil {
		 fmt.Printf("❌ ACL认证失败: %v\n", err)
		 return
	 }
	 fmt.Println("✅ ACL认证成功")
 
	 client.Close()
 }
 
 func testTLSConnection() {
	 fmt.Println("\n=== 测试3: TLS连接 ===")
	 
	 // 检查证书文件是否存在
	 if _, err := os.Stat("certs/ca.crt"); os.IsNotExist(err) {
		 fmt.Println("⚠️  TLS证书文件不存在，跳过TLS测试")
		 fmt.Println("   请先运行: ./scripts/generate-certs.sh")
		 return
	 }
 
	 tlsConfig := &tls.Config{
		 MinVersion: tls.VersionTLS12,
	 }
 
	 // 加载CA证书
	 caCert, err := os.ReadFile("certs/ca.crt")
	 if err != nil {
		 fmt.Printf("❌ 读取CA证书失败: %v\n", err)
		 return
	 }
 
	 caCertPool := x509.NewCertPool()
	 if !caCertPool.AppendCertsFromPEM(caCert) {
		 fmt.Println("❌ 解析CA证书失败")
		 return
	 }
	 tlsConfig.RootCAs = caCertPool
 
	 client := redis.NewClient(&redis.Options{
		 Addr:      "127.0.0.1:6380", // TLS端口
		 Password:  "",
		 DB:        0,
		 TLSConfig: tlsConfig,
	 })
 
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()
 
	 _, err = client.Ping(ctx).Result()
	 if err != nil {
		 fmt.Printf("❌ TLS连接失败: %v\n", err)
		 return
	 }
	 fmt.Println("✅ TLS连接成功")
 
	 client.Close()
 }
 
 func testACLAndTLS() {
	 fmt.Println("\n=== 测试4: 组合使用ACL和TLS ===")
	 
	 // 检查证书文件是否存在
	 if _, err := os.Stat("certs/ca.crt"); os.IsNotExist(err) {
		 fmt.Println("⚠️  TLS证书文件不存在，跳过组合测试")
		 return
	 }
 
	 tlsConfig := &tls.Config{
		 MinVersion: tls.VersionTLS12,
	 }
 
	 // 加载CA证书
	 caCert, err := os.ReadFile("certs/ca.crt")
	 if err != nil {
		 fmt.Printf("❌ 读取CA证书失败: %v\n", err)
		 return
	 }
 
	 caCertPool := x509.NewCertPool()
	 if !caCertPool.AppendCertsFromPEM(caCert) {
		 fmt.Println("❌ 解析CA证书失败")
		 return
	 }
	 tlsConfig.RootCAs = caCertPool
 
	 client := redis.NewClient(&redis.Options{
		 Addr:      "127.0.0.1:6380", // TLS端口
		 Username:  "benchmark",
		 Password:  "mypassword",
		 DB:        0,
		 TLSConfig: tlsConfig,
	 })
 
	 ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	 defer cancel()
 
	 _, err = client.Ping(ctx).Result()
	 if err != nil {
		 fmt.Printf("❌ ACL+TLS组合测试失败: %v\n", err)
		 return
	 }
	 fmt.Println("✅ ACL+TLS组合测试成功")
 
	 client.Close()
 }
 
 func init() {
	 log.SetFlags(log.LstdFlags | log.Lshortfile)
 } 