# 项目优化方案（lesson-plan）

## 1. 项目整体结构分析

### 1.1 架构概览

本项目采用「前端 + 后端 + Agent + 多数据库」分层架构：

- `frontend/`：Vue3 + Vite + Pinia，负责交互与页面渲染
- `backend/`：Go + Gin + GORM，负责业务 API 与权限
- `agent/`：TypeScript + LangGraph，负责教案智能生成与知识检索
- `database/`：PostgreSQL / Neo4j 初始化脚本
- 根目录脚本：`start.sh` / `stop.sh` / `start-win.ps1` / `stop-win.ps1` 用于一键启动

### 1.2 现状优点

- 技术栈清晰，目录职责明确，便于协作分工
- 本地与容器两种运行模式均已具备
- 环境变量已在 README 中约定“统一放根目录 .env”
- Windows 启动脚本已有较完整的健康等待逻辑

### 1.3 关键问题（可优先落地）

1. **Linux/macOS 启动脚本误报后端失败**
   - `start.sh` 仅等待 3 秒后检查端口，依赖服务（尤其 Neo4j）稍慢即误判。

2. **后端 CORS 中间件存在逻辑缺陷**
   - 先按 `Origin` 设置头，后又用 `strings.Join(allowedOrigins)` 覆盖；
   - 当 `Allow-Credentials=true` 且允许 `*` 时不符合浏览器 CORS 约束。

3. **路由层未使用配置文件中的 CORS / 限流参数**
   - 当前硬编码 `NewCORSMiddleware([]string{"*"}, true)` 和 `NewRateLimitMiddleware(100, 200)`，配置项形同虚设。

4. **开发者体验细节**
   - `.env.example` 缺失会影响新同学初始化；
   - 优化方案缺少落地文档，知识无法沉淀。

---

## 2. 优化目标

### 目标 A：提升启动稳定性与可观测性
- 减少“误报失败”，明确依赖未就绪时的错误原因。

### 目标 B：修复后端跨域策略正确性
- 满足 CORS 规范，避免浏览器环境下偶发请求失败。

### 目标 C：让配置真正生效
- 统一走 `config.yaml + .env`，减少硬编码差异。

### 目标 D：补齐交付文档
- 形成可持续维护的优化记录和环境模板。

---

## 3. 逐条实施清单（含结果）

### 3.1 启动脚本可靠性优化（已实施）

**改动文件**：`start.sh`

实施内容：
- 增加 `docker compose` / `docker-compose` 自动探测；
- 新增依赖与服务等待函数：
  - `wait_for_port`（端口就绪检测）
  - `wait_for_http`（HTTP 健康检查）
- 容器启动后显式等待：`5432`、`17687`、`6379`；
- 后端启动改为“最长等待 + /health 检查”，不再固定 `sleep 3`；
- Agent 启动改为超时等待，不再固定 `sleep 5`；
- 增加可配置参数：
  - `BACKEND_PORT`
  - `BACKEND_STARTUP_TIMEOUT`
  - `AGENT_STARTUP_TIMEOUT`
  - `DOCKER_STARTUP_TIMEOUT`

预期收益：
- 显著减少“后端失败”的假告警；
- 依赖没起来时定位更直接。

### 3.2 CORS 中间件正确性修复（已实施）

**改动文件**：`backend/internal/middleware/cors.go`

实施内容：
- 删除“二次覆盖 `Access-Control-Allow-Origin`”的问题；
- 引入 `isOriginAllowed`，只在允许时回写当前 `Origin`；
- `AllowCredentials=true && AllowOrigins=*` 时改为回显请求源并添加 `Vary: Origin`；
- `Max-Age` 改为读取配置值，不再硬编码 `86400`。

预期收益：
- 与浏览器 CORS 行为一致；
- 降低预检请求失败概率。

### 3.3 路由层接入配置项（已实施）

**改动文件**：
- `backend/internal/handler/router.go`
- `backend/cmd/server/main.go`

实施内容：
- `Router` 注入 `*config.Config`；
- 路由初始化时读取 `cfg.CORS` 和 `cfg.RateLimit`；
- 仅在 `rate_limit.enabled=true` 时注册限流中间件；
- `main.go` 调整 `NewRouter(...)` 参数，传入 `cfg`。

预期收益：
- 运行行为与配置文件一致；
- 环境切换时无需改代码。

### 3.4 补齐环境模板（已实施）

**改动文件**：`.env.example`

实施内容：
- 恢复并补全根目录环境模板；
- 同步新增启动脚本可选超时变量。

预期收益：
- 新环境初始化更顺畅；
- 减少“变量漏配”排查成本。

---

## 4. 后续建议（下一阶段）

1. **统一停止脚本行为**
   - `stop.sh` 当前会交互询问是否关闭容器，建议提供 `--with-docker` 参数以便 CI/自动化调用。

2. **补充后端基础测试**
   - 至少为 `cors.go` 与 `config` 覆盖关键行为，防止回归。

3. **根目录增加统一质量命令**
   - 例如 `make lint-all` / `make build-all` / `make check`，减少跨目录手工切换。

4. **清理误提交产物的风险**
   - 建议继续保持 `dist/`、`node_modules/` 忽略策略，并在 PR 流程增加检查。

---

## 5. 验证建议

可按以下顺序快速验证：

1. `./start.sh`：观察后端不再误报失败；
2. `curl http://localhost:8080/health`：确认健康检查可用；
3. 浏览器打开前端并执行登录/业务接口，确认 CORS 正常；
4. 调整 `backend/config/config.yaml` 的 `cors` / `rate_limit`，重启后验证配置生效。

