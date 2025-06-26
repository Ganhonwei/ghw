#!/bin/bash

set -e

# 项目根目录
PROJECT_DIR=$(cd $(dirname $0)/..; pwd)

# 配置表目录
TABLE_DIR="${PROJECT_DIR}/deps/table"

# 检查必要目录
if [[ ! -d ${TABLE_DIR} ]]; then
    echo "配置表目录不存在"
    exit 1
fi

# 进入配置表目录
cd ${TABLE_DIR}

# 生成配置表代码及数据
./gen.sh

# 进入配置表数据目录  
cd "${TABLE_DIR}/Datas"

# 复制json配置数据
mkdir -p "${PROJECT_DIR}/configs/free" && cp -f -r *.json "${PROJECT_DIR}/configs/free/"

