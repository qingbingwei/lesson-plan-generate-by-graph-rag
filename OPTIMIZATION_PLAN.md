# 教案生成系统 - 优化计划

## 项目概览

- **Agent**: TypeScript + Express.js + LangGraph (AI教案生成工作流)
- **Backend**: Go + Gin + GORM (PostgreSQL) + Neo4j + Redis
- **Frontend**: Vue 3 + Pinia + Vue Router + Tailwind CSS + D3.js + Vite

---

## 一、安全优化 (S1-S10)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| S1 | Agent 服务无认证，任何人可调用 | `agent/src/index.ts` | 🔴 高 |
| S2 | 请求体缺少 Schema 验证 (Zod/Joi) | `agent/src/index.ts` 各路由 | 🔴 高 |
| S3 | JWT Secret 硬编码在配置文件中 | `backend/config/config.yaml` | 🔴 高 |
| S4 | 密码策略缺失 (无复杂度要求) | `backend/internal/service/user_service.go` | 🟡 中 |
| S5 | Rate Limiter 使用内存存储 (多实例不共享) | `backend/internal/middleware/ratelimit.go` | 🟡 中 |
| S6 | CORS 允许所有来源 | `backend/internal/middleware/cors.go` | 🟡 中 |
| S7 | 文件上传无类型/大小白名单校验 | `backend/internal/handler/knowledge_handler.go` | 🔴 高 |
| S8 | 知识图谱查询字符串拼接 (Cypher注入) | `agent/src/tools/neo4j.ts` | 🔴 高 |
| S9 | 前端Token存储在localStorage (XSS风险) | `frontend/src/stores/auth.ts` | 🟡 中 |
| S10 | 缺少 CSP (Content Security Policy) | `frontend/nginx.conf` | 🟡 中 |

## 二、架构优化 (A1-A12) ✅ 已执行

| 编号 | 问题 | 位置 | 状态 |
|------|------|------|------|
| A1 | Agent index.ts 单体文件 (336行) | `agent/src/index.ts` | ✅ 已完成 |
| A2 | 重复代码：mergeUsage × 3, SkillSchema × 4, normalizeGrade × 2, getClient × 3 | `agent/src/nodes/`, `agent/src/skills/` | ✅ 已完成 |
| A3 | 死代码：buildSkillsPrompt、loadSkill 从未调用 | `agent/src/skills/index.ts` | ✅ 已完成 |
| A4 | 工作流缺少错误短路 (每个节点手动检查 error) | `agent/src/workflow/lessonWorkflow.ts` | ✅ 已完成 |
| A5 | Go后端全局变量管理DB连接 | `backend/pkg/database/*.go` | ✅ 已完成 |
| A6 | DocumentRepository 无接口 (其他Repo均有) | `backend/internal/repository/document_repository.go` | ✅ 已完成 |
| A7 | lesson_service.go (532行) 和 generation_service.go (495行) 包含多个服务 | `backend/internal/service/` | ✅ 已完成 |
| A8 | goroutine 无 recover/context/超时 | `backend/internal/service/document_service.go` | ✅ 已完成 |
| A9 | 前端 generation store 两个函数 ~100行重复 | `frontend/src/stores/generation.ts` | ✅ 已完成 |
| A10 | API 模块风格不一致 (knowledge用对象字面量, 其他用函数导出) | `frontend/src/api/knowledge.ts` | ✅ 已完成 |
| A11 | 自定义 composables 重复 @vueuse/core 功能 | `frontend/src/composables/index.ts` | ✅ 已完成 |
| A12 | localStorage 有3套抽象 (composable + utils + pinia-plugin) | `frontend/src/composables/` + `frontend/src/utils/` | ✅ 已完成 |

## 三、性能优化 (P1-P8)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| P1 | LLM 调用串行执行 (可并行的环节未并行) | `agent/src/nodes/*.ts` | 🔴 高 |
| P2 | 知识图谱查询无缓存 | `agent/src/tools/neo4j.ts` | 🟡 中 |
| P3 | 前端无路由懒加载 | `frontend/src/router/index.ts` | 🟡 中 |
| P4 | D3 图谱组件无虚拟化 (大数据量卡顿) | `frontend/src/components/` | 🟡 中 |
| P5 | 数据库查询无索引优化声明 | `database/postgres/init.sql` | 🟡 中 |
| P6 | Redis helpers 全部未使用 | `backend/pkg/database/redis.go` | 🟢 低 |
| P7 | 大文件处理无分片/流式 | `backend/internal/service/document_service.go` | 🟡 中 |
| P8 | 前端 bundle 未分析优化 | `frontend/vite.config.ts` | 🟢 低 |

## 四、错误处理优化 (E1-E8)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| E1 | Agent 节点错误仅日志不恢复 | `agent/src/nodes/*.ts` | 🟡 中 |
| E2 | LLM 调用无重试机制 | `agent/src/clients/deepseek.ts` | 🔴 高 |
| E3 | SSE 流断开无重连 | `agent/src/index.ts` (stream route) | 🟡 中 |
| E4 | 后端 `_ = err` 大量忽略错误 | `backend/internal/service/lesson_service.go` | 🟡 中 |
| E5 | 前端缺少全局错误边界 | `frontend/src/App.vue` | 🟡 中 |
| E6 | API 响应无统一错误码体系 | `backend/internal/handler/response.go` | 🟡 中 |
| E7 | Neo4j 连接无健康检查/自动重连 | `agent/src/tools/neo4j.ts` | 🟡 中 |
| E8 | 前端请求无全局Loading/错误提示 | `frontend/src/api/index.ts` | 🟢 低 |

## 五、用户体验优化 (U1-U10)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| U1 | 教案生成进度为假动画 (8秒间隔纯前端模拟) | `frontend/src/stores/generation.ts` | ✅ 已完成 |
| U2 | 知识图谱可视化交互有限 | `frontend/src/views/Knowledge.vue` | ✅ 已完成 |
| U3 | 无教案导出功能 (PDF/Word/Markdown) | `backend/internal/handler/lesson_handler.go` | ✅ 已有 |
| U4 | 缺少教案版本对比/历史 | 全局 | 🟢 低 (需后端 schema) |
| U5 | 移动端适配不完整 | `frontend/src/views/Knowledge.vue` 等 | ✅ 已完成 |
| U6 | 暗色模式实现不完整 | `frontend/src/composables/index.ts` | 🟢 低 (需全局 CSS) |
| U7 | 无批量生成/模板功能 | 全局 | 🟢 低 (需后端) |
| U8 | 搜索无防抖和高亮 | `frontend/src/views/Lessons.vue` 等 | ✅ 已完成 |
| U9 | 文档上传无进度条 | `frontend/src/api/knowledge.ts` | ✅ 已完成 |
| U10 | 无操作引导/新手教程 | `frontend/src/views/Dashboard.vue` | ✅ 已完成 |

## 六、基础设施优化 (I1-I8)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| I1 | 无 CI/CD 配置 | 项目根目录 | 🔴 高 |
| I2 | Docker镜像无多阶段构建优化 | `*/Dockerfile` | 🟡 中 |
| I3 | 无健康检查端点 (Docker compose) | `docker-compose.yml` | 🟡 中 |
| I4 | 无日志聚合方案 | 全局 | 🟡 中 |
| I5 | 无监控/告警 | 全局 | 🟡 中 |
| I6 | 数据库无备份策略 | `docker-compose.yml` | 🔴 高 |
| I7 | 环境变量管理不规范 | `docker-compose.yml` + 各组件 | 🟡 中 |
| I8 | Makefile 缺少测试/lint 目标 | `Makefile` | 🟢 低 |

## 七、代码质量优化 (Q1-Q7)

| 编号 | 问题 | 位置 | 优先级 |
|------|------|------|--------|
| Q1 | Agent 缺少单元测试 | `agent/` | 🔴 高 |
| Q2 | Backend 缺少单元测试 | `backend/` | 🔴 高 |
| Q3 | Frontend 缺少组件测试 | `frontend/` | 🟡 中 |
| Q4 | 缺少 API 文档 (Swagger/OpenAPI) | `backend/` | 🟡 中 |
| Q5 | TypeScript 类型定义不严格 (`as any` 多处) | `agent/src/workflow/*.ts` | 🟡 中 |
| Q6 | 缺少 ESLint/Prettier 统一配置 | 各组件 | 🟢 低 |
| Q7 | 未使用依赖 (dayjs, @types/marked 放在dependencies) | `frontend/package.json` | 🟢 低 |

---

## 执行优先级

1. **🔴 安全类** (S1-S3, S7-S8) - 立即修复
2. **🔴 架构类** (A1-A12) - ✅ 已执行
3. **🔴 核心性能/错误** (P1, E2) - 下一批
4. **🟡 中等优先** (其他) - 迭代改进
5. **🟢 低优先** - 长期优化
