#!/bin/bash

# 教案生成系统启动脚本
# 启动顺序：后端 -> Agent -> 前端

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp/lesson-plan"

# 创建日志目录
mkdir -p "$LOG_DIR"

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  教案生成系统 - 启动服务${NC}"
echo -e "${GREEN}========================================${NC}"

# 检查Docker容器是否运行
echo -e "\n${YELLOW}[1/4] 检查 Docker 容器...${NC}"
if ! docker ps | grep -q "lesson-plan-postgres"; then
    echo -e "${YELLOW}启动 Docker 容器...${NC}"
    cd "$PROJECT_DIR" && docker-compose up -d
    sleep 5
else
    echo -e "${GREEN}✓ Docker 容器已在运行${NC}"
fi

# 启动后端服务
echo -e "\n${YELLOW}[2/4] 启动后端服务 (Go)...${NC}"
cd "$PROJECT_DIR/backend"
if lsof -ti :8080 > /dev/null 2>&1; then
    echo -e "${YELLOW}端口 8080 已被占用，正在关闭...${NC}"
    kill $(lsof -ti :8080) 2>/dev/null || true
    sleep 2
fi
nohup go run ./cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "后端 PID: $BACKEND_PID"

# 等待后端启动
sleep 3
if lsof -ti :8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ 后端服务已启动 (http://localhost:8080)${NC}"
else
    echo -e "${RED}✗ 后端服务启动失败，查看日志: $LOG_DIR/backend.log${NC}"
fi

# 启动Agent服务
echo -e "\n${YELLOW}[3/4] 启动 Agent 服务 (Node.js)...${NC}"
cd "$PROJECT_DIR/agent"
if lsof -ti :3001 > /dev/null 2>&1; then
    echo -e "${YELLOW}端口 3001 已被占用，正在关闭...${NC}"
    kill $(lsof -ti :3001) 2>/dev/null || true
    sleep 2
fi
nohup npm run dev > "$LOG_DIR/agent.log" 2>&1 &
AGENT_PID=$!
echo "Agent PID: $AGENT_PID"

# 等待Agent启动
sleep 5
if lsof -ti :3001 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Agent 服务已启动 (http://localhost:3001)${NC}"
else
    echo -e "${RED}✗ Agent 服务启动失败，查看日志: $LOG_DIR/agent.log${NC}"
fi

# 启动前端服务
echo -e "\n${YELLOW}[4/4] 启动前端服务 (Vite)...${NC}"
cd "$PROJECT_DIR/frontend"
# 关闭可能存在的前端进程
pkill -f "vite.*lesson-plan" 2>/dev/null || true
sleep 1
nohup npm run dev > "$LOG_DIR/frontend.log" 2>&1 &
FRONTEND_PID=$!
echo "前端 PID: $FRONTEND_PID"

# 等待前端启动
sleep 3
FRONTEND_PORT=$(grep -o "localhost:[0-9]*" "$LOG_DIR/frontend.log" 2>/dev/null | head -1 | cut -d: -f2)
if [ -n "$FRONTEND_PORT" ]; then
    echo -e "${GREEN}✓ 前端服务已启动 (http://localhost:$FRONTEND_PORT)${NC}"
else
    echo -e "${YELLOW}前端服务正在启动中...${NC}"
fi

# 保存PID到文件
echo "$BACKEND_PID" > "$LOG_DIR/backend.pid"
echo "$AGENT_PID" > "$LOG_DIR/agent.pid"
echo "$FRONTEND_PID" > "$LOG_DIR/frontend.pid"

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  所有服务启动完成！${NC}"
echo -e "${GREEN}========================================${NC}"
echo -e "\n服务地址："
echo -e "  前端:  ${GREEN}http://localhost:${FRONTEND_PORT:-5173}${NC}"
echo -e "  后端:  ${GREEN}http://localhost:8080${NC}"
echo -e "  Agent: ${GREEN}http://localhost:3001${NC}"
echo -e "\n日志目录: $LOG_DIR"
echo -e "  - backend.log"
echo -e "  - agent.log"
echo -e "  - frontend.log"
echo -e "\n使用 ${YELLOW}./stop.sh${NC} 关闭所有服务"
