#!/bin/bash
#
# go-redisbenchmark - A high performance Redis benchmark tool written in Go.
# Supports Redis 6.0+ features including ACL and TLS.
#
# Copyright (c) 2025, dal1wg <15072697283@139.com>
# All rights reserved.
#
# This source code is licensed under the BSD 3-Clause License.
#

# 生成TLS证书脚本
# 用于Redis 6.0+ TLS测试

set -e

CERT_DIR="certs"
mkdir -p $CERT_DIR

echo "生成CA私钥..."
openssl genrsa -out $CERT_DIR/ca.key 4096

echo "生成CA证书..."
openssl req -new -x509 -days 365 -key $CERT_DIR/ca.key -out $CERT_DIR/ca.crt -subj "/C=CN/ST=Beijing/L=Beijing/O=RedisTest/OU=IT/CN=RedisTestCA"

echo "生成Redis服务器私钥..."
openssl genrsa -out $CERT_DIR/redis.key 2048

echo "生成Redis服务器证书签名请求..."
openssl req -new -key $CERT_DIR/redis.key -out $CERT_DIR/redis.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=RedisTest/OU=IT/CN=localhost"

echo "生成Redis服务器证书..."
openssl x509 -req -days 365 -in $CERT_DIR/redis.csr -CA $CERT_DIR/ca.crt -CAkey $CERT_DIR/ca.key -CAcreateserial -out $CERT_DIR/redis.crt

echo "生成客户端私钥..."
openssl genrsa -out $CERT_DIR/client.key 2048

echo "生成客户端证书签名请求..."
openssl req -new -key $CERT_DIR/client.key -out $CERT_DIR/client.csr -subj "/C=CN/ST=Beijing/L=Beijing/O=RedisTest/OU=IT/CN=redis-client"

echo "生成客户端证书..."
openssl x509 -req -days 365 -in $CERT_DIR/client.csr -CA $CERT_DIR/ca.crt -CAkey $CERT_DIR/ca.key -CAcreateserial -out $CERT_DIR/client.crt

echo "清理临时文件..."
rm -f $CERT_DIR/*.csr $CERT_DIR/*.srl

echo "设置权限..."
chmod 600 $CERT_DIR/*.key
chmod 644 $CERT_DIR/*.crt

echo "TLS证书生成完成！"
echo "证书文件位置: $CERT_DIR/"
echo "  - ca.crt: CA证书"
echo "  - redis.key: Redis服务器私钥"
echo "  - redis.crt: Redis服务器证书"
echo "  - client.key: 客户端私钥"
echo "  - client.crt: 客户端证书" 