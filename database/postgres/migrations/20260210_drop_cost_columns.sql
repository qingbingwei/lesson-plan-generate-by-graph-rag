-- 移除历史 cost 列（仅保留 token 使用量）
BEGIN;

ALTER TABLE IF EXISTS generations
  DROP COLUMN IF EXISTS cost;

ALTER TABLE IF EXISTS generation_logs
  DROP COLUMN IF EXISTS cost;

COMMIT;
