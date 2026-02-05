# 智能教案生成系统

基于 AI 的智能教案生成系统，支持知识图谱构建、Graph RAG 检索和自动化教案生成。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | Vue 3 + Vite + Pinia + Tailwind CSS |
| 后端 | Go + Gin + GORM |
| 智能体 | LangGraph + TypeScript + DeepSeek API |
| 数据库 | PostgreSQL + Neo4j + Redis |

## 快速开始

### 环境要求
- Docker & Docker Compose
- DeepSeek API Key

### 一键启动

```bash
# 1. 配置环境变量
cp .env.example .env
# 编辑 .env 填入 DEEPSEEK_API_KEY

# 2. 启动所有服务
./start.sh
```

启动后访问：
- 前端：http://localhost:5173
- 后端 API：http://localhost:8080
- Neo4j Browser：http://localhost:7474

### 常用命令

```bash
make up          # 启动服务
make down        # 停止服务
make logs        # 查看日志
make logs-backend   # 后端日志
make logs-agent     # 智能体日志
make ps          # 服务状态
make clean       # 清理数据（慎用）
```

## 项目结构

```
lesson-plan/
├── frontend/          # Vue 3 前端
├── backend/           # Go 后端 API
├── agent/             # LangGraph 智能体
├── database/          # 数据库初始化脚本
├── docker-compose.yml
├── start.sh           # 一键启动脚本
└── Makefile           # 常用命令
```

## 核心功能

1. **教案生成** - AI 自动生成结构化教案（目标、内容、活动、评价）
2. **知识图谱** - 上传文档自动构建个人知识图谱
3. **Graph RAG** - 基于知识图谱的智能检索增强生成
4. **教案管理** - 编辑、收藏、导出（Markdown/Word/PDF）

## API 概览

```
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
