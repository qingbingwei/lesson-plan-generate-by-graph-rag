#!/bin/bash

# 教案生成系统启动脚本
# 启动顺序：后端 -> Agent -> 前端

set -e

PROJECT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_DIR="/tmp/lesson-plan"
ENV_FILE="$PROJECT_DIR/.env"

if [ -f "$ENV_FILE" ]; then
    set -a
    # shellcheck disable=SC1090
    . "$ENV_FILE"
    set +a
fi

AGENT_PORT="${AGENT_PORT:-${PORT:-13001}}"
BACKEND_PORT="${BACKEND_PORT:-8080}"
BACKEND_STARTUP_TIMEOUT="${BACKEND_STARTUP_TIMEOUT:-30}"
AGENT_STARTUP_TIMEOUT="${AGENT_STARTUP_TIMEOUT:-30}"
DOCKER_STARTUP_TIMEOUT="${DOCKER_STARTUP_TIMEOUT:-60}"

DOCKER_COMPOSE_CMD=()

resolve_docker_compose_cmd() {
    if docker compose version >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD=(docker compose)
        return 0
    fi

    if command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD=(docker-compose)
        return 0
    fi

    return 1
}

run_compose() {
    "${DOCKER_COMPOSE_CMD[@]}" "$@"
}

is_container_running() {
    local name="$1"
    docker ps --format '{{.Names}}' | grep -q "^${name}$"
}

wait_for_port() {
    local port="$1"
    local timeout="$2"
    local elapsed=0

    while [ "$elapsed" -lt "$timeout" ]; do
        if lsof -ti ":${port}" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done

    return 1
}

wait_for_http() {
    local url="$1"
    local timeout="$2"
    local elapsed=0

    if ! command -v curl >/dev/null 2>&1; then
        return 1
    fi

    while [ "$elapsed" -lt "$timeout" ]; do
        if curl -fsS --max-time 2 "$url" >/dev/null 2>&1; then
            return 0
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done

    return 1
}

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

if [ -f "$PROJECT_DIR/agent/.env" ]; then
    echo -e "${YELLOW}[WARN] 检测到 agent/.env，当前已统一使用根目录 .env，请删除 agent/.env 避免混淆${NC}"
fi

# 检查Docker容器是否运行
echo -e "\n${YELLOW}[1/4] 检查 Docker 容器...${NC}"
if ! command -v docker >/dev/null 2>&1; then
    echo -e "${RED}✗ 未检测到 docker 命令，请先安装 Docker Desktop${NC}"
    exit 1
fi

if ! docker info >/dev/null 2>&1; then
    echo -e "${RED}✗ Docker 未启动或当前账号无权限访问 Docker${NC}"
    exit 1
fi

if ! resolve_docker_compose_cmd; then
    echo -e "${RED}✗ 未检测到 docker compose / docker-compose${NC}"
    exit 1
fi

if ! is_container_running "lesson-plan-postgres" || \
   ! is_container_running "lesson-plan-neo4j" || \
   ! is_container_running "lesson-plan-redis"; then
    echo -e "${YELLOW}启动 Docker 容器...${NC}"
    cd "$PROJECT_DIR" && run_compose up -d postgres neo4j redis
else
    echo -e "${GREEN}✓ Docker 容器已在运行${NC}"
fi

echo -e "${YELLOW}等待依赖服务就绪...${NC}"
if ! wait_for_port 5432 "$DOCKER_STARTUP_TIMEOUT"; then
    echo -e "${RED}✗ PostgreSQL 未在 ${DOCKER_STARTUP_TIMEOUT}s 内就绪 (5432)${NC}"
    exit 1
fi
if ! wait_for_port 17687 "$DOCKER_STARTUP_TIMEOUT"; then
    echo -e "${RED}✗ Neo4j 未在 ${DOCKER_STARTUP_TIMEOUT}s 内就绪 (17687)${NC}"
    exit 1
fi
if ! wait_for_port 6379 "$DOCKER_STARTUP_TIMEOUT"; then
    echo -e "${RED}✗ Redis 未在 ${DOCKER_STARTUP_TIMEOUT}s 内就绪 (6379)${NC}"
    exit 1
fi

# 启动后端服务
echo -e "\n${YELLOW}[2/4] 启动后端服务 (Go)...${NC}"
cd "$PROJECT_DIR/backend"
if lsof -ti :$BACKEND_PORT > /dev/null 2>&1; then
    echo -e "${YELLOW}端口 $BACKEND_PORT 已被占用，正在关闭...${NC}"
    kill $(lsof -ti :$BACKEND_PORT) 2>/dev/null || true
    sleep 2
fi
nohup go run ./cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
BACKEND_PID=$!
echo "后端 PID: $BACKEND_PID"

# 等待后端启动
if wait_for_port "$BACKEND_PORT" "$BACKEND_STARTUP_TIMEOUT"; then
    if wait_for_http "http://localhost:${BACKEND_PORT}/health" 5; then
        echo -e "${GREEN}✓ 后端服务已启动 (http://localhost:${BACKEND_PORT})${NC}"
    else
        echo -e "${YELLOW}✓ 后端端口已打开，但健康检查未在 5s 内通过${NC}"
        echo -e "${YELLOW}  请查看日志: $LOG_DIR/backend.log${NC}"
    fi
else
    echo -e "${RED}✗ 后端服务启动失败，查看日志: $LOG_DIR/backend.log${NC}"
fi

# 启动Agent服务
echo -e "\n${YELLOW}[3/4] 启动 Agent 服务 (Node.js)...${NC}"
cd "$PROJECT_DIR/agent"
if lsof -ti :$AGENT_PORT > /dev/null 2>&1; then
    echo -e "${YELLOW}端口 $AGENT_PORT 已被占用，正在关闭...${NC}"
    kill $(lsof -ti :$AGENT_PORT) 2>/dev/null || true
    sleep 2
fi
if [ "$AGENT_PORT" != "3001" ] && lsof -ti :3001 > /dev/null 2>&1; then
    kill $(lsof -ti :3001) 2>/dev/null || true
    sleep 1
fi
AGENT_PORT="$AGENT_PORT" PORT="$AGENT_PORT" nohup npm run dev > "$LOG_DIR/agent.log" 2>&1 &
AGENT_PID=$!
echo "Agent PID: $AGENT_PID"

# 等待Agent启动
if wait_for_port "$AGENT_PORT" "$AGENT_STARTUP_TIMEOUT"; then
    echo -e "${GREEN}✓ Agent 服务已启动 (http://localhost:$AGENT_PORT)${NC}"
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
echo -e "  后端:  ${GREEN}http://localhost:$BACKEND_PORT${NC}"
echo -e "  Agent: ${GREEN}http://localhost:$AGENT_PORT${NC}"
echo -e "\n日志目录: $LOG_DIR"
echo -e "  - backend.log"
echo -e "  - agent.log"
echo -e "  - frontend.log"
echo -e "\n使用 ${YELLOW}./stop.sh${NC} 关闭所有服务"
