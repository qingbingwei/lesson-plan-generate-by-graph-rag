# Schema Changelog

用于记录所有数据库结构变更的审计信息。

## 记录模板

| Date (UTC) | Migration File | Type | Objects | Forward Result | Rollback Result | Owner | Reviewer | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-02-20T00:00:00Z | 20260220143000_example.sql | DDL | users, idx_users_email | success | success | @owner | @reviewer | 示例记录 |

## 历史记录

| Date (UTC) | Migration File | Type | Objects | Forward Result | Rollback Result | Owner | Reviewer | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-02-10T00:00:00Z | 20260210_drop_cost_columns.sql | DDL | generations.cost, generation_logs.cost | success | pending (未演练) | team-backend | pending | 移除冗余 cost 字段，仅保留 token 使用量 |

