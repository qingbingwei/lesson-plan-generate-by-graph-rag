include .env
export

.PHONY: help
help: ## 显示帮助信息
	@echo "可用命令："
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: init
init: ## 初始化项目（创建.env文件）
	@if [ ! -f .env ]; then \
		cp .env.example .env; \
		echo "✅ 已创建 .env 文件，请编辑并填入必要的配置"; \
	else \
		echo "⚠️  .env 文件已存在"; \
	fi

.PHONY: up
up: ## 启动所有服务
	docker-compose up -d

.PHONY: down
down: ## 停止所有服务
	docker-compose down

.PHONY: restart
restart: down up ## 重启所有服务

.PHONY: logs
logs: ## 查看所有服务日志
	docker-compose logs -f

.PHONY: logs-backend
logs-backend: ## 查看后端日志
	docker-compose logs -f backend

.PHONY: logs-agent
logs-agent: ## 查看智能体日志
	docker-compose logs -f agent

.PHONY: logs-frontend
logs-frontend: ## 查看前端日志
	docker-compose logs -f frontend

.PHONY: ps
ps: ## 查看服务状态
	docker-compose ps

.PHONY: clean
clean: ## 清理所有容器和数据卷（⚠️ 危险操作）
	docker-compose down -v
	@echo "⚠️  所有数据已删除"

.PHONY: init-db
init-db: ## 初始化数据库
	@echo "正在初始化PostgreSQL..."
	docker-compose exec -T postgres psql -U admin -d lesson_plan < database/postgres/init.sql
	@echo "正在初始化Neo4j..."
	docker-compose exec -T neo4j cypher-shell -u neo4j -p $(NEO4J_PASSWORD) -f /var/lib/neo4j/import/init.cypher
	@echo "✅ 数据库初始化完成"

.PHONY: seed-db
seed-db: ## 导入样例数据
	@echo "正在导入样例数据..."
	docker-compose exec -T postgres psql -U admin -d lesson_plan < database/postgres/seed.sql
	@echo "✅ 样例数据导入完成"

.PHONY: backup-db
backup-db: ## 备份数据库
	@mkdir -p backups
	@echo "备份PostgreSQL..."
	docker-compose exec -T postgres pg_dump -U admin lesson_plan > backups/postgres_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "✅ 数据库备份完成"

.PHONY: dev-backend
dev-backend: ## 本地开发模式启动后端
	cd backend && go run cmd/server/main.go

.PHONY: dev-agent
dev-agent: ## 本地开发模式启动智能体
	cd agent && npm run dev

.PHONY: dev-frontend
dev-frontend: ## 本地开发模式启动前端
	cd frontend && npm run dev

.PHONY: test-backend
test-backend: ## 运行后端测试
	cd backend && go test ./... -v

.PHONY: test-agent
test-agent: ## 运行智能体测试
	cd agent && npm test

.PHONY: test-frontend
test-frontend: ## 运行前端测试
	cd frontend && npm test

.PHONY: build-backend
build-backend: ## 构建后端
	cd backend && go build -o bin/server cmd/server/main.go

.PHONY: build-agent
build-agent: ## 构建智能体
	cd agent && npm run build

.PHONY: build-frontend
build-frontend: ## 构建前端
	cd frontend && npm run build

.PHONY: docker-build
docker-build: ## 构建所有Docker镜像
	docker-compose build

.PHONY: install-backend
install-backend: ## 安装后端依赖
	cd backend && go mod download

.PHONY: install-agent
install-agent: ## 安装智能体依赖
	cd agent && npm install

.PHONY: install-frontend
install-frontend: ## 安装前端依赖
	cd frontend && npm install

.PHONY: install
install: install-backend install-agent install-frontend ## 安装所有依赖

.PHONY: format-backend
format-backend: ## 格式化后端代码
	cd backend && go fmt ./...

.PHONY: format-agent
format-agent: ## 格式化智能体代码
	cd agent && npm run format

.PHONY: format-frontend
format-frontend: ## 格式化前端代码
	cd frontend && npm run format

.PHONY: lint-backend
lint-backend: ## 检查后端代码
	cd backend && golangci-lint run

.PHONY: lint-agent
lint-agent: ## 检查智能体代码
	cd agent && npm run lint

.PHONY: lint-frontend
lint-frontend: ## 检查前端代码
	cd frontend && npm run lint

.PHONY: shell-postgres
shell-postgres: ## 进入PostgreSQL命令行
	docker-compose exec postgres psql -U admin -d lesson_plan

.PHONY: shell-neo4j
shell-neo4j: ## 进入Neo4j命令行
	docker-compose exec neo4j cypher-shell -u neo4j -p $(NEO4J_PASSWORD)

.PHONY: shell-redis
shell-redis: ## 进入Redis命令行
	docker-compose exec redis redis-cli -a $(REDIS_PASSWORD)
