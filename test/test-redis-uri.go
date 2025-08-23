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
	 "fmt"
	 "log"
	 "goRedisBenchmark/pkg"
 )
 
 func main() {
	 fmt.Println("测试Redis URI解析功能...")
 
	 testCases := []struct {
		 name        string
		 uri         string
		 description string
	 }{
		 {
			 name:        "基本ACL认证",
			 uri:         "redis://benchmark:mypassword@127.0.0.1:6379/0",
			 description: "用户名:benchmark, 密码:mypassword, 主机:127.0.0.1:6379, 数据库:0",
		 },
		 {
			 name:        "TLS连接",
			 uri:         "rediss://127.0.0.1:6380/0",
			 description: "TLS连接, 主机:127.0.0.1:6380, 数据库:0",
		 },
		 {
			 name:        "ACL+TLS组合",
			 uri:         "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true",
			 description: "ACL认证+TLS连接",
		 },
		 {
			 name:        "完整TLS配置",
			 uri:         "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-cert=client.crt&tls-key=client.key&tls-ca=ca.crt",
			 description: "ACL认证+完整TLS配置",
		 },
		 {
			 name:        "跳过TLS验证",
			 uri:         "redis://benchmark:mypassword@127.0.0.1:6380/0?tls=true&tls-insecure=true",
			 description: "ACL认证+TLS连接+跳过验证",
		 },
		 {
			 name:        "无认证连接",
			 uri:         "redis://127.0.0.1:6379/1",
			 description: "无认证, 主机:127.0.0.1:6379, 数据库:1",
		 },
	 }
 
	 for _, tc := range testCases {
		 fmt.Printf("\n=== %s ===\n", tc.name)
		 fmt.Printf("URI: %s\n", tc.uri)
		 fmt.Printf("描述: %s\n", tc.description)
		 
		 opts := benchmark.Options{
			 URI: tc.uri,
		 }
		 
		 runner, err := benchmark.NewRunner(opts)
		 if err != nil {
			 fmt.Printf("❌ 创建Runner失败: %v\n", err)
			 continue
		 }
		 
		 // 测试URI解析
		 if err := runner.ParseRedisURI(); err != nil {
			 fmt.Printf("❌ URI解析失败: %v\n", err)
			 continue
		 }
		 
		 fmt.Printf("✅ URI解析成功\n")
		 fmt.Printf("   地址: %s\n", opts.Addr)
		 fmt.Printf("   数据库: %d\n", opts.DB)
		 fmt.Printf("   用户名: %s\n", opts.Username)
		 fmt.Printf("   ACL密码: %s\n", opts.ACLPassword)
		 fmt.Printf("   使用TLS: %t\n", opts.UseTLS)
		 fmt.Printf("   TLS证书: %s\n", opts.TLSCertFile)
		 fmt.Printf("   TLS私钥: %s\n", opts.TLSKeyFile)
		 fmt.Printf("   TLS CA: %s\n", opts.TLSCAFile)
		 fmt.Printf("   TLS跳过验证: %t\n", opts.TLSInsecure)
	 }
 }
 
 func init() {
	 log.SetFlags(log.LstdFlags | log.Lshortfile)
 } 