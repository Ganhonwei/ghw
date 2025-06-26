#!/bin/bash

set -e

# 项目根目录
PROJECT_DIR=$(cd $(dirname $0)/..; pwd)

# 协议文件目录
PROTO_DIR="${PROJECT_DIR}/deps/proto"

# 协议生成目录
PB_DIR="${PROJECT_DIR}/gen/pb"

# 命令码生成目录
CMD_PB_DIR="${PROJECT_DIR}/misc/tool/pb"

# 工具目录
TOOL_DIR="${PROJECT_DIR}/misc/tool"

# 检查必要目录
if [[ ! -d ${PROTO_DIR} ]]; then
    echo "协议文件不存在"
    exit 1
fi

# 定义协议文件列表
PROTO_FILES="actor*.proto game*.proto hua*.proto lhd*.proto up*.proto \
            crash*.proto rm*.proto joker*.proto ak47*.proto andar*.proto \
            lottery*.proto plane*.proto redblack*.proto mines*.proto \
            fortune_gems2*.proto fortune_gems*.proto"

# 进入协议目录
cd ${PROTO_DIR}

# 生成协议
protoc -I=. -I=${PROJECT_DIR}/third_party/ --gogoslick_out=plugins=grpc:${PB_DIR} ${PROTO_FILES}

# 生成命令码
protoc -I=. --gogoslick_out=plugins=grpc:${CMD_PB_DIR} const_commands.proto

# 进入工具目录
cd ${TOOL_DIR}

# 生成打包和解包代码
go run gen.go