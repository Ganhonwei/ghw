#!/bin/bash

set -e

# 基础配置
DOCKERFILE="build/global/dockerfile"
REGISTRY="reg.mydomain.com"
NAMESPACE="library-test"

IMAGE_NAME="global"
BUILD_TYPE="release"

FULL_IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}"

# 显示使用方法
usage() {
    echo "Usage: ./build.sh <version> [--debug] [--ns <namespace>]"
    echo "  <version>    : 版本号(必需)"
    echo "  --debug      : 使用debug版本(可选)"
    echo "  --ns <value> : 指定命名空间(可选，默认: library-test)"
    exit 1
}

# 解析参数
parse_args() {
    if [ $# -lt 1 ]; then
        usage
    fi
    
    export TAG=$1
    shift  # 移除第一个参数（版本号）
    
    # 检查其他参数
    while [ "$1" != "" ]; do
        case $1 in
            --debug )   
                        IMAGE_NAME="global-debug"
                        BUILD_TYPE="debug"
                        ;;
            --ns ) 
                        shift
                        NAMESPACE=$1
                        ;;
            * )         
                        usage
                        ;;
        esac
        shift
    done
    
    # 更新完整镜像名
    FULL_IMAGE_NAME="${REGISTRY}/${NAMESPACE}/${IMAGE_NAME}"
}

# 检查 TAG 参数
check_tag() {
    if [ -z "$TAG" ]; then
        echo "TAG is required."
        usage
    fi
}

# 构建 Docker 镜像
build_image() {
    echo "Using Dockerfile: ${DOCKERFILE}"
    echo "Image name: ${FULL_IMAGE_NAME}:${TAG}"
    
    docker build \
        --build-arg BUILD_TYPE=${BUILD_TYPE} \
        --target ${IMAGE_NAME} \
        --tag ${FULL_IMAGE_NAME}:${TAG} \
        -f ${DOCKERFILE} .
}

# 推送 Docker 镜像
push_image() {
    docker push ${FULL_IMAGE_NAME}:${TAG}
}

# 主函数
main() {
    parse_args "$@"
    check_tag
    build_image
    push_image
}

# 执行脚本
main "$@"