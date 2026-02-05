#!/bin/bash

# 教案生成系统关闭脚本

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp/lesson-plan"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  教案生成系统 - 关闭服务${NC}"
echo -e "${YELLOW}========================================${NC}"

# 关闭前端服务
echo -e "\n${YELLOW}[1/3] 关闭前端服务...${NC}"
pkill -f "vite" 2>/dev/null && echo -e "${GREEN}✓ 前端服务已关闭${NC}" || echo -e "${YELLOW}前端服务未运行${NC}"

# 关闭Agent服务
echo -e "\n${YELLOW}[2/3] 关闭 Agent 服务...${NC}"
if lsof -ti :3001 > /dev/null 2>&1; then
    kill $(lsof -ti :3001) 2>/dev/null
    echo -e "${GREEN}✓ Agent 服务已关闭${NC}"
else
    echo -e "${YELLOW}Agent 服务未运行${NC}"
fi

# 关闭后端服务
echo -e "\n${YELLOW}[3/3] 关闭后端服务...${NC}"
if lsof -ti :8080 > /dev/null 2>&1; then
    kill $(lsof -ti :8080) 2>/dev/null
    echo -e "${GREEN}✓ 后端服务已关闭${NC}"
else
    echo -e "${YELLOW}后端服务未运行${NC}"
fi

# 清理残留的npm和node进程
pkill -f "npm run dev" 2>/dev/null || true
pkill -f "ts-node" 2>/dev/null || true
pkill -f "tsx" 2>/dev/null || true

# 清理PID文件
rm -f "$LOG_DIR"/*.pid 2>/dev/null

sleep 1

# 验证关闭
echo -e "\n${YELLOW}验证服务状态...${NC}"
STILL_RUNNING=0

if lsof -ti :8080 > /dev/null 2>&1; then
    echo -e "${RED}✗ 后端服务仍在运行 (端口 8080)${NC}"
    STILL_RUNNING=1
fi

if lsof -ti :3001 > /dev/null 2>&1; then
    echo -e "${RED}✗ Agent 服务仍在运行 (端口 3001)${NC}"
    STILL_RUNNING=1
fi

if pgrep -f "vite" > /dev/null 2>&1; then
    echo -e "${RED}✗ 前端服务仍在运行${NC}"
    STILL_RUNNING=1
fi

if [ $STILL_RUNNING -eq 0 ]; then
    echo -e "${GREEN}✓ 所有服务已成功关闭${NC}"
fi

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  完成${NC}"
echo -e "${GREEN}========================================${NC}"

# 询问是否关闭Docker容器
echo -e "\n是否同时关闭 Docker 容器? (y/N)"
read -r response
if [[ "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo -e "${YELLOW}关闭 Docker 容器...${NC}"
    cd "$PROJECT_DIR" && docker-compose down
    echo -e "${GREEN}✓ Docker 容器已关闭${NC}"
fi
