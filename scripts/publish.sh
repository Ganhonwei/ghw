#!/bin/bash

set -e

# 项目根目录
PROJECT_DIR=$(cd $(dirname $0)/..; pwd)

# 发布目录
PUBLISH_DIR="${PROJECT_DIR}/publish/config"

# 配置目录
CONFIG_DIR="${PROJECT_DIR}/configs"

# 检查参数
if [ $# -ne 1 ]; then
    echo "Usage: ./publish.sh <version>"
    echo "Example: ./publish.sh v1.0.0"
    exit 1
fi

VERSION=$1

# 进入发布目录
cd ${PUBLISH_DIR} || {
    echo "Error: publish directory not found"
    exit 1
}

# 切换git版本
git checkout $VERSION || {
    echo "Error: Failed to checkout version $VERSION"
    exit 1
}

# 进入配置目录
cd ${CONFIG_DIR} || {
    echo "Error: config directory not found"
    exit 1
}

# 拷贝free目录到publish/config
cp -r free ${PUBLISH_DIR}/ || {
    echo "Error: Failed to copy free directory"
    exit 1
}

# 拷贝table目录到publish/config
cp -r table ${PUBLISH_DIR}/ || {
    echo "Error: Failed to copy table directory"
    exit 1
}

echo "- Successfully published version $VERSION"
echo "- Copied configs/free to publish/config/free"
echo "- Copied configs/table to publish/config/table"
