#!/bin/bash
# 校验内嵌在 Go 二进制里的前端 JS 语法。
# 背景：ui.go 里的 HTML/JS 是字符串，Go 编译器不会检查内部语法，
#       一个多余的 } 就能让整个列表页白屏，且编译、探针 HTTP 200 都发现不了。
#
# 用法： bash scripts/check-embedded-js.sh <index.cgi 路径>
# 依赖： node（可选，没有则只做括号平衡检查）
set -u
CGI="${1:-app/ui/index.cgi}"

if [ ! -f "$CGI" ]; then
  echo "找不到二进制: $CGI" >&2
  exit 1
fi

P="$CGI.probe-js"
OUT=$(mktemp /tmp/page-XXXX.html)
PJ=$(mktemp /tmp/page-XXXX.js)
trap 'rm -f "$P" "$OUT" "$PJ"' EXIT

cp "$CGI" "$P"; chmod 755 "$P"
printf '' | env REQUEST_METHOD=GET \
  REQUEST_URI=/cgi/ThirdParty/iconManager/index.cgi/ \
  QUERY_STRING= GATEWAY_INTERFACE=CGI/1.1 "$P" 2>/dev/null > "$OUT"

SIZE=$(wc -c < "$OUT")
echo "页面大小: $SIZE 字节"
if [ "$SIZE" -lt 1000 ]; then
  echo "❌ 页面异常小，CGI 可能未正常运行" >&2
  exit 1
fi

python3 - "$OUT" > "$PJ" <<'PY'
import re, sys
html = open(sys.argv[1], encoding='utf-8', errors='replace').read()
blocks = re.findall(r'<script[^>]*>(.*?)</script>', html, re.S)
sys.stdout.write(max(blocks, key=len) if blocks else '')
PY

JSIZE=$(wc -c < "$PJ")
echo "内嵌 JS: $JSIZE 字节"
if [ "$JSIZE" -lt 100 ]; then
  echo "❌ 未提取到 JS" >&2
  exit 1
fi

# 括号平衡（不依赖 node）
L=$(grep -o '{' "$PJ" | wc -l)
R=$(grep -o '}' "$PJ" | wc -l)
if [ "$L" != "$R" ]; then
  echo "❌ 花括号不平衡: { =$L  } =$R" >&2
  exit 1
fi
echo "花括号平衡: { =$L  } =$R"

if command -v node >/dev/null 2>&1; then
  if node --check "$PJ" 2>/tmp/node-err-$$; then
    echo "✅ JS 语法检查通过"
  else
    echo "❌ JS 语法错误:" >&2
    cat /tmp/node-err-$$ >&2
    rm -f /tmp/node-err-$$
    exit 1
  fi
  rm -f /tmp/node-err-$$
else
  echo "（无 node，仅做了括号平衡检查）"
fi
