# 智能教案生成系统

基于 AI 的智能教案生成系统，支持知识图谱构建、Graph RAG 检索和自动化教案生成。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vite + Pinia + Element Plus + Tailwind CSS |
| 后端 | Go + Gin + GORM |
| 智能体 | LangGraph + TypeScript + DeepSeek API |
| 数据库 | PostgreSQL + Neo4j + Redis |

## 快速开始

### 环境要求
- Docker Desktop（含 Docker Compose）
- Node.js（建议 20+）
- Go（建议 1.22+）
- DeepSeek API Key

## 一键启动（Linux/macOS）

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入 DEEPSEEK_API_KEY

# 2. 启动所有服务
./start.sh
```

> `start.sh` 已支持更稳健的依赖就绪检测与后端健康检查（`/health`）。
> 如需调整等待时间，可在 `.env` 里设置：
>
> - `BACKEND_STARTUP_TIMEOUT=30`
> - `AGENT_STARTUP_TIMEOUT=30`
> - `DOCKER_STARTUP_TIMEOUT=60`

## 一键启动（Windows）

```powershell
# 1) 启动依赖容器（postgres / neo4j / redis）
docker compose up -d postgres neo4j redis

# 2) 启动本地后端 + Agent + 前端
.\start-win.ps1
# 或
.\start.bat
```

> 若遇到 Agent 端口冲突，统一在根目录 `.env` 设置：
>
> - `AGENT_PORT=13001`（主配置）
> - `PORT=13001`（兼容旧代码）
>
> `start-win.ps1` 会优先读取 `AGENT_PORT`，并自动清理旧的 `3001` Agent 进程。

停止：

```powershell
.\stop-win.ps1
# 或
.\stop.bat

# 连同容器一起关闭
.\stop-win.ps1 -StopDocker
```

## 访问地址

- 前端：`http://localhost:5173`
- 后端 API：`http://localhost:8080`
- Agent：`http://localhost:13001`
- Neo4j Browser：`http://localhost:17474`

> 说明：Windows 下为了避开系统保留端口，Neo4j 端口映射为 `17474:7474`、`17687:7687`。

## 环境变量规范（推荐）

统一原则：
- 所有环境变量只放在项目根目录 `.env`
- 不在 `agent/.env`、`frontend/.env`、`backend/.env` 放重复配置

- `VITE_API_BASE_URL=/api/v1`
- `VITE_BACKEND_PROXY_TARGET=http://localhost:8080`
- `AGENT_PORT=13001`
- `AGENT_SERVICE_URL=http://localhost:13001`
- `VITE_AGENT_BASE_URL=http://localhost:13001`
- `QWEN_API_KEY=...`
- `QWEN_EMBEDDING_MODEL=text-embedding-v4`
- `QWEN_EMBEDDING_URL=https://dashscope.aliyuncs.com/compatible-mode/v1/embeddings`

说明：
- 前端已兼容多种 `VITE_API_BASE_URL` 写法（`/api/v1`、`http://localhost:8080`、`http://localhost:8080/api` 等），会自动归一化到正确的 `/api/v1` 路径，避免注册登录出现 `404`。
- Agent 与启动脚本统一读取根目录 `.env`，`AGENT_PORT` 为主，`PORT` 仅兼容保留。
- Qwen Embedding 配置同样统一放在根目录 `.env`，不要再在 `agent/.env` 里单独维护。

### LangSmith 可视化分析

Agent 已支持 LangSmith tracing（LangGraph/LangChain 运行轨迹可视化）：

- `LANGSMITH_TRACING=true`
- `LANGSMITH_API_KEY=...`
- `LANGSMITH_ENDPOINT=https://api.smith.langchain.com`
- `LANGSMITH_PROJECT=lesson-plan-agent`

说明：
- 当 `LANGSMITH_TRACING=true` 且 `LANGSMITH_API_KEY` 有值时，Agent 会自动开启 tracing。
- 教案生成调用会带上 `runName/tags/metadata`，便于在 LangSmith 中筛选与分析。
- 若开启 tracing 但未配置 API Key，服务会打印告警并自动关闭 tracing，避免影响主流程。

## 常用命令

```bash
make up            # 启动服务
make down          # 停止服务
make logs          # 查看日志
make logs-backend  # 后端日志
make logs-agent    # Agent 日志
make ps            # 服务状态
make clean         # 清理数据（慎用）
```

## 优化方案文档

- 结构分析与已实施优化：`OPTIMIZATION_PLAN.md`

## 项目结构

```text
lesson-plan/
├── frontend/          # Vue 3 前端
├── backend/           # Go 后端 API
├── agent/             # LangGraph 智能体
├── database/          # 数据库初始化脚本
├── docker-compose.yml
├── start.sh           # Linux/macOS 启动脚本
├── stop.sh            # Linux/macOS 停止脚本
├── start-win.ps1      # Windows 启动脚本
├── stop-win.ps1       # Windows 停止脚本
└── Makefile
```

## 核心功能

1. **教案生成**：AI 自动生成结构化教案（目标、内容、活动、评价）
2. **知识图谱**：上传文档后自动构建个人知识图谱
3. **Graph RAG**：基于知识图谱的检索增强生成
4. **教案管理**：编辑、收藏、导出（Markdown/Word/PDF）

## API 概览

```text
POST /api/v1/auth/register    # 注册
POST /api/v1/auth/login       # 登录
POST /api/v1/generate         # 生成教案
GET  /api/v1/lessons          # 教案列表
GET  /api/v1/lessons/:id      # 教案详情
PUT  /api/v1/lessons/:id      # 更新教案
GET  /api/v1/knowledge/graph  # 知识图谱
POST /api/v1/knowledge/upload # 上传文档
```

## License

MIT
