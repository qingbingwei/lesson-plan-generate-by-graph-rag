-- Migration: <YYYYMMDDHHMMSS_action_target>
-- Author: <name>
-- Date(UTC): <YYYY-MM-DD>
-- Description: <what and why>
-- Risk: <low|medium|high>
-- Notes: <optional>

BEGIN;

-- [FORWARD]
-- 1) 在这里填写上线 SQL
-- 2) 建议使用 IF EXISTS / IF NOT EXISTS 保证幂等
-- 3) 如果涉及大表，请在注释中说明执行窗口与预估耗时


-- [ROLLBACK]
-- 1) 在这里填写回滚 SQL
-- 2) 回滚必须覆盖本次 FORWARD 的关键结构变更
-- 3) 如果无法完全回滚，必须明确声明并给出替代方案


COMMIT;

