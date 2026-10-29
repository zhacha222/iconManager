#!/bin/bash
# 把图标数据从源应用迁移到 iconManager 名下。
#
# 原理：
#   - app_service 表无外键约束，app_id 只是普通字段
#   - 图标 URL 形如 /app-center-static/serviceicon/<app_name>/<icon>，第一段是 app_name
#   - 因此迁移 = 改 app_id + 拷贝图标文件到新应用目录
#
# 用法：
#   bash migrate.sh check     仅检查，不改任何东西
#   bash migrate.sh backup    备份当前状态
#   bash migrate.sh run       执行迁移
#   bash migrate.sh verify    迁移后验证
#   bash migrate.sh rollback  回滚到迁移前
#
# 可用环境变量覆盖：
#   PG_DSN          PostgreSQL 连接串
#   SRC_APP_ID      源应用 id（默认自动查 customIcon）
#   DST_APP_NAME    目标应用名（默认 iconManager）
#   BACKUP_DIR      备份目录（默认 /var/backups/iconManager）
set -uo pipefail

PG_DSN="${PG_DSN:-host=/var/run/postgresql user=postgres dbname=appcenter sslmode=disable}"
DST_APP_NAME="${DST_APP_NAME:-iconManager}"
BACKUP_DIR="${BACKUP_DIR:-/var/backups/iconManager}"

q() { psql "$PG_DSN" -A -t -c "$1"; }
qq() { psql "$PG_DSN" -q -c "$1"; }

# 自动解析 app_id 与应用路径
resolve_ids() {
  SRC_APP_ID="${SRC_APP_ID:-$(q "SELECT id FROM app WHERE app_name='customIcon';")}"
  DST_APP_ID="$(q "SELECT id FROM app WHERE app_name='$DST_APP_NAME';")"
  SRC_APP_NAME="$(q "SELECT app_name FROM app WHERE id=$SRC_APP_ID;")"
  SRC_PATH="$(q "SELECT path FROM app WHERE id=$SRC_APP_ID;")"
  DST_PATH="$(q "SELECT path FROM app WHERE id=$DST_APP_ID;")"
  SRC_IMAGES="${SRC_PATH%/}/target/ui/images"
  DST_IMAGES="${DST_PATH%/}/target/ui/images"

  if [ -z "$SRC_APP_ID" ] || [ -z "$DST_APP_ID" ]; then
    echo "错误：无法解析应用 id（源=$SRC_APP_ID 目标=$DST_APP_ID）" >&2
    exit 1
  fi
}

cmd_check() {
  resolve_ids
  echo "=== 源应用 ==="
  q "SELECT id||' | '||app_name||' | '||name||' | '||status FROM app WHERE id=$SRC_APP_ID;"
  echo "=== 目标应用 ==="
  q "SELECT id||' | '||app_name||' | '||name||' | '||status FROM app WHERE id=$DST_APP_ID;"
  echo
  echo "=== 待迁移图标数 ==="
  q "SELECT count(*) FROM app_service WHERE app_id=$SRC_APP_ID;"
  echo
  echo "=== 源图标目录 ==="
  echo "  $SRC_IMAGES ($(ls "$SRC_IMAGES" 2>/dev/null | wc -l) 个文件)"
  echo "=== 目标图标目录 ==="
  echo "  $DST_IMAGES"
  echo
  echo "=== 外键约束（应为 0，否则改 app_id 可能有风险）==="
  q "SELECT count(*) FROM pg_constraint WHERE conrelid='app_service'::regclass AND contype='f';"
}

cmd_backup() {
  mkdir -p "$BACKUP_DIR"
  local ts; ts=$(date +%Y%m%d-%H%M%S)
  echo "$ts" > "$BACKUP_DIR/.last-migrate-stamp"
  q "SELECT json_agg(row_to_json(t)) FROM (SELECT * FROM app_service ORDER BY id) t;" \
    > "$BACKUP_DIR/app_service-$ts.json"
  q "SELECT json_agg(row_to_json(t)) FROM (SELECT * FROM app ORDER BY id) t;" \
    > "$BACKUP_DIR/app-$ts.json"
  echo "已备份到 $BACKUP_DIR/*-$ts.json"
  ls -la "$BACKUP_DIR"/*"$ts".json
}

cmd_run() {
  resolve_ids
  if [ ! -f "$BACKUP_DIR/.last-migrate-stamp" ]; then
    echo "错误：请先执行 backup" >&2; exit 1
  fi

  echo "=== 1. 创建目标图标目录 ==="
  mkdir -p "$DST_IMAGES"

  echo "=== 2. 拷贝图标文件 ==="
  local n=0
  for f in "$SRC_IMAGES"/*; do
    [ -f "$f" ] || continue
    cp -n "$f" "$DST_IMAGES/$(basename "$f")" && n=$((n+1))
  done
  chmod 644 "$DST_IMAGES"/* 2>/dev/null
  echo "  拷贝 $n 个，目标现有 $(ls "$DST_IMAGES" | wc -l) 个"

  echo "=== 3. 改 app_id（$SRC_APP_ID → $DST_APP_ID）==="
  qq "UPDATE app_service SET app_id=$DST_APP_ID WHERE app_id=$SRC_APP_ID;" | tail -1

  echo "=== 4. 结果 ==="
  q "SELECT 'app_id='||app_id||' 共 '||count(*)||' 条' FROM app_service
     WHERE app_id IN ($SRC_APP_ID,$DST_APP_ID) GROUP BY app_id;"
}

cmd_verify() {
  resolve_ids
  echo "=== 迁移结果 ==="
  q "SELECT 'app_id='||app_id||' 共 '||count(*)||' 条' FROM app_service
     WHERE app_id IN ($SRC_APP_ID,$DST_APP_ID) GROUP BY app_id;"
  echo
  echo "=== 图标文件检查 ==="
  local miss=0
  while IFS= read -r ic; do
    [ -z "$ic" ] && continue
    local base; base=$(basename "$ic")
    case "$base" in *"{0}"*) continue;; esac
    if [ ! -f "$DST_IMAGES/$base" ]; then
      echo "  缺失: $base"; miss=$((miss+1))
    fi
  done < <(q "SELECT icon FROM app_service WHERE app_id=$DST_APP_ID;")
  [ "$miss" = 0 ] && echo "  全部图标文件就位" || echo "  共 $miss 个缺失"
}

cmd_rollback() {
  resolve_ids
  if [ ! -f "$BACKUP_DIR/.last-migrate-stamp" ]; then
    echo "错误：找不到备份时间戳" >&2; exit 1
  fi
  local ts; ts=$(cat "$BACKUP_DIR/.last-migrate-stamp")
  local f="$BACKUP_DIR/app_service-$ts.json"
  [ -s "$f" ] || { echo "错误：备份文件不存在 $f" >&2; exit 1; }

  echo "=== 从 $f 恢复 ==="
  python3 - "$f" <<'PY' | psql "$PG_DSN" -q
import json, sys
rows = json.load(sys.stdin) if False else json.load(open(sys.argv[1]))
print("BEGIN;")
print("TRUNCATE app_service;")
for r in rows:
    cols = ','.join('"%s"' % k for k in r.keys())
    def val(v):
        if v is None: return 'NULL'
        if isinstance(v, bool): return 'true' if v else 'false'
        if isinstance(v, (int, float)): return str(v)
        return "'" + str(v).replace("'", "''") + "'"
    print("INSERT INTO app_service (%s) VALUES (%s);" % (cols, ','.join(val(v) for v in r.values())))
print("COMMIT;")
PY
  echo "=== 恢复后 ==="
  q "SELECT 'app_id='||app_id||' 共 '||count(*)||' 条' FROM app_service
     WHERE app_id IN ($SRC_APP_ID,$DST_APP_ID) GROUP BY app_id;"
}

case "${1:-}" in
  check)    cmd_check ;;
  backup)   cmd_backup ;;
  run)      cmd_run ;;
  verify)   cmd_verify ;;
  rollback) cmd_rollback ;;
  *) echo "用法: bash migrate.sh {check|backup|run|verify|rollback}"; exit 1 ;;
esac
