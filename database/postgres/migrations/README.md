# PostgreSQL Migration 规范

本目录用于管理所有 PostgreSQL 结构变更（DDL）与数据修复脚本（必要时 DML）。

## 强制规则

1. 所有 schema 变更必须通过 migration 落地，禁止直接在数据库手工改表结构。
2. 每个 migration 文件必须同时提供前滚（forward）与回滚（rollback）步骤。
3. 每次合并 migration 时，必须同步更新 `SCHEMA_CHANGELOG.md`。
4. migration 执行必须放在事务中；如果语句不支持事务，需要在文件头明确标注原因。

## 文件命名

统一命名格式：

`YYYYMMDDHHMMSS_<action>_<target>.sql`

示例：

- `20260220143000_add_user_profile_indexes.sql`
- `20260220150000_alter_generations_add_trace_id.sql`

## 编写要求

1. 基于 `TEMPLATE.sql` 新建 migration。
2. `-- [FORWARD]` 段只放上线变更所需 SQL。
3. `-- [ROLLBACK]` 段必须可在紧急情况下执行回退。
4. 涉及大表时，需在注释中标明风险与执行窗口建议。
5. 涉及数据修复时，需给出幂等策略（`IF EXISTS`/`IF NOT EXISTS`/条件更新）。

## 变更流程

1. 新建 migration 文件并补齐 forward/rollback。
2. 在 `SCHEMA_CHANGELOG.md` 追加一条记录（变更目的、影响范围、回滚方式、审批人）。
3. 在测试环境执行前滚与回滚演练并记录结果。
4. 通过 CI 后再进入生产发布流程。

## 审计要求

`SCHEMA_CHANGELOG.md` 至少记录以下信息：

- 变更时间（UTC）
- migration 文件名
- 变更类型（DDL/DML）
- 影响对象（表/索引/约束）
- 前滚结果
- 回滚验证结果
- 责任人 / 评审人

