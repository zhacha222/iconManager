// iconManager —— 飞牛扩展图标管理（CGI 模式，对齐官方 customIcon）
//
// 运行方式：单文件可执行，飞牛 trim_http_cgi 每次请求时启动一次，处理完退出。
// 不监听端口、无常驻进程，空闲零占用 —— 与官方 customIcon 一致。
//
// 路由：/cgi/ThirdParty/iconManager/index.cgi/<path>
//   /                      → HTML 页面
//   /api/icons             → 列表
//   /api/icon?f=...        → 图标文件
//   /api/service?name=...  → 单条
//   /api/{create,update,delete,disable,enable,upload,reload}
package main

import (
	"bytes"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

const (
	dsn           = "host=/var/run/postgresql user=postgres dbname=appcenter sslmode=disable"
	selfAppName   = "iconManager"
	serviceUnit   = "trim_app_center.service"

	defaultControl = `{"show":0,"showRoute":0,"auth":0,"port":1,"path":1,"fullUrl":1,` +
		`"accessPerm":"editable","portPerm":"readonly","pathPerm":"readonly","fullUrlPerm":"editable"}`

)

var (
	db        *sql.DB
	selfAppID int64
	logf      *log.Logger
	reloadMu  sync.Mutex
)

type Icon struct {
	ID          int64  `json:"id"`
	ServiceName string `json:"service_name"`
	Title       string `json:"title"`
	Icon        string `json:"icon"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	IsAdmin     bool   `json:"is_admin"`
	NoDisplay   bool   `json:"no_display"`
	Disabled    bool   `json:"disabled"`
	IconURL     string `json:"icon_url"`
}

type BackupRecord struct {
	ServiceName string          `json:"service_name"`
	DisabledAt  string          `json:"disabled_at"`
	Row         json.RawMessage `json:"row"`
}

func main() {
	// 日志：CGI 下每次请求都会重写一行，用 append 模式
	pkgVar := os.Getenv("TRIM_PKGVAR")
	if pkgVar == "" {
		pkgVar = "/tmp/iconManager"
	}
	os.MkdirAll(pkgVar, 0o755)
	if lf, err := os.OpenFile(filepath.Join(pkgVar, "cgi.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
		logf = log.New(lf, "", log.LstdFlags)
	} else {
		logf = log.New(os.Stderr, "", log.LstdFlags)
	}

	// 若在终端里直接运行（非 CGI 环境），提示用法
	if os.Getenv("REQUEST_METHOD") == "" && os.Getenv("GATEWAY_INTERFACE") == "" {
		fmt.Fprintln(os.Stderr, "这是飞牛 CGI 程序，需由 trim_http_cgi 调用。")
		fmt.Fprintln(os.Stderr, "访问：https://<NAS域名>/cgi/ThirdParty/iconManager/index.cgi/")
		os.Exit(1)
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		cgiFail("打开数据库失败: " + err.Error())
		return
	}
	defer db.Close()
	db.SetMaxOpenConns(2)
	if err := db.Ping(); err != nil {
		cgiFail("数据库连接失败: " + err.Error())
		return
	}

	// app_id 不能写死（卸载重装会变）
	if err := db.QueryRow(`SELECT id FROM app WHERE app_name=$1`, selfAppName).Scan(&selfAppID); err != nil {
		cgiFail("无法解析本应用 app_id: " + err.Error())
		return
	}

	serveCGI()
}

// cgiFail 在 CGI 环境下输出错误页并结束
func cgiFail(msg string) {
	logf.Printf("ERROR: %s", msg)
	fmt.Println("Status: 500 Internal Server Error")
	fmt.Println("Content-Type: text/html; charset=utf-8")
	fmt.Println()
	fmt.Printf("<h2>iconManager 出错</h2><pre>%s</pre>", msg)
}

// serveCGI 解析 CGI 环境变量并派发
func serveCGI() {
	// 从 REQUEST_URI 或 PATH_INFO 取出 index.cgi 之后的路径
	uri := os.Getenv("REQUEST_URI")
	if uri == "" {
		uri = os.Getenv("PATH_INFO")
	}
	if i := strings.Index(uri, "?"); i >= 0 {
		uri = uri[:i]
	}
	path := "/"
	if i := strings.Index(uri, "index.cgi"); i >= 0 {
		path = uri[i+len("index.cgi"):]
		if path == "" {
			path = "/"
		}
	} else if pi := os.Getenv("PATH_INFO"); pi != "" {
		path = pi
	}

	method := os.Getenv("REQUEST_METHOD")
	qs := os.Getenv("QUERY_STRING")

	body, _ := io.ReadAll(os.Stdin)

	req := struct {
		Method string
		Path   string
		Query  string
		Body   []byte
	}{method, path, qs, body}

	logf.Printf("%s %s", req.Method, req.Path)

	// 管理员判定：CGI 由飞牛在调用前校验 NAS 登录态
	// 通过环境变量或请求头拿不到用户信息时，视为已登录（内网 + NAS 鉴权已生效）
	isAdmin := true
	if v := os.Getenv("HTTP_X_TRIM_ISADMIN"); v != "" {
		isAdmin = v == "true"
	}

	switch {
	case req.Path == "/" || req.Path == "":
		writeHTML()
	case req.Path == "/api/icons":
		apiList(isAdmin)
	case req.Path == "/api/icon":
		apiIconFile(req.Query)
	case req.Path == "/api/service":
		apiService(req.Query)
	case req.Path == "/api/create":
		apiCreate(req.Body, isAdmin)
	case req.Path == "/api/update":
		apiUpdate(req.Body, isAdmin)
	case req.Path == "/api/delete":
		apiDelete(req.Body, isAdmin)
	case req.Path == "/api/disable":
		apiDisable(req.Body, isAdmin)
	case req.Path == "/api/enable":
		apiEnable(req.Body, isAdmin)
	case req.Path == "/api/upload":
		apiUpload(req.Body, isAdmin)
	case req.Path == "/api/reload":
		apiReload(isAdmin)
	case req.Path == "/api/healthz":
		outJSON(200, map[string]string{"status": "ok"})
	default:
		outJSON(404, map[string]string{"error": "not found: " + req.Path})
	}
}

func writeHTML() {
	fmt.Println("Content-Type: text/html; charset=utf-8")
	fmt.Println("Cache-Control: no-store")
	fmt.Println()
	fmt.Print(indexHTML)
}

func outJSON(code int, v any) {
	b, _ := json.Marshal(v)
	fmt.Printf("Status: %d\n", code)
	fmt.Println("Content-Type: application/json; charset=utf-8")
	fmt.Println("Cache-Control: no-store")
	fmt.Println()
	fmt.Print(string(b))
}

func outErr(code int, msg string) { outJSON(code, map[string]string{"error": msg}) }

// ---------- 工具 ----------

func esc(s string) string {
	r := strings.NewReplacer("/", "_", "\\", "_", " ", "_", "..", "_")
	return r.Replace(s)
}

func backupRoot() string {
	if v := os.Getenv("TRIM_PKGVAR"); v != "" {
		return filepath.Join(v, "backups")
	}
	return "/tmp/iconManager/backups"
}

func backupDir(name string) string { return filepath.Join(backupRoot(), esc(name)) }

// imagesDir 返回图标存放目录。优先级：环境变量 > TRIM_APPDEST 推导 > 默认值。
// 不能硬编码 /vol2：不同设备卷号可能是 vol1、vol3 等。
func imagesDir() string {
	if v := os.Getenv("ICONMGR_IMAGES_DIR"); v != "" {
		return v
	}
	if dest := os.Getenv("TRIM_APPDEST"); dest != "" {
		return filepath.Join(dest, "ui", "images")
	}
	return "/var/apps/iconManager/target/ui/images"
}

func iconURL(icon string) string {
	return "/api/icon?f=" + strings.ReplaceAll(strings.TrimSpace(icon), "{0}", "256")
}

// ---------- 查询 ----------

func queryIcons() ([]Icon, error) {
	rows, err := db.Query(`
		SELECT id, service_name, coalesce(title,''), coalesce(icon,''),
		       coalesce("type",''), coalesce(url,''), is_admin, no_display
		FROM app_service WHERE app_id = $1 ORDER BY id`, selfAppID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]Icon, 0, 32)
	for rows.Next() {
		var ic Icon
		if err := rows.Scan(&ic.ID, &ic.ServiceName, &ic.Title, &ic.Icon,
			&ic.Type, &ic.URL, &ic.IsAdmin, &ic.NoDisplay); err != nil {
			return nil, err
		}
		ic.IconURL = iconURL(ic.Icon)
		out = append(out, ic)
	}
	return out, rows.Err()
}

func loadDisabled() []Icon {
	root := backupRoot()
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil
	}
	out := []Icon{}
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "_") {
			continue
		}
		blob, err := os.ReadFile(filepath.Join(root, e.Name(), "record.json"))
		if err != nil {
			continue
		}
		var rec BackupRecord
		if json.Unmarshal(blob, &rec) != nil || rec.ServiceName == "" {
			continue
		}
		var row struct {
			ID        int64  `json:"id"`
			Title     string `json:"title"`
			Icon      string `json:"icon"`
			Type      string `json:"type"`
			URL       string `json:"url"`
			IsAdmin   bool   `json:"is_admin"`
			NoDisplay bool   `json:"no_display"`
		}
		_ = json.Unmarshal(rec.Row, &row)
		out = append(out, Icon{
			ID: row.ID, ServiceName: rec.ServiceName, Title: row.Title,
			Icon: row.Icon, Type: row.Type, URL: row.URL,
			IsAdmin: row.IsAdmin, NoDisplay: row.NoDisplay,
			Disabled: true, IconURL: iconURL(row.Icon),
		})
	}
	return out
}

// ---------- 图片处理 ----------

func resizeTo(src image.Image, w, h int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	if sw == 0 || sh == 0 {
		return dst
	}
	for y := 0; y < h; y++ {
		sy := sb.Min.Y + y*sh/h
		for x := 0; x < w; x++ {
			sx := sb.Min.X + x*sw/w
			dst.Set(x, y, src.At(sx, sy))
		}
	}
	return dst
}

func saveIconSizes(baseName string, data []byte) (string, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("无法解析图片: %w", err)
	}
	os.MkdirAll(imagesDir(), 0o755)
	for _, sz := range []int{256, 64} {
		out := resizeTo(src, sz, sz)
		canvas := image.NewRGBA(image.Rect(0, 0, sz, sz))
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
		draw.Draw(canvas, canvas.Bounds(), out, image.Point{}, draw.Over)
		p := filepath.Join(imagesDir(), fmt.Sprintf("%s_%d.png", baseName, sz))
		f, err := os.Create(p)
		if err != nil {
			return "", err
		}
		err = png.Encode(f, canvas)
		f.Close()
		if err != nil {
			return "", err
		}
		os.Chmod(p, 0o644)
	}
	return fmt.Sprintf("%s_256.png", baseName), nil
}

// ---------- API ----------

func apiList(isAdmin bool) {
	icons, err := queryIcons()
	if err != nil {
		outErr(500, "查询失败: "+err.Error())
		return
	}
	live := map[string]bool{}
	for _, ic := range icons {
		live[ic.ServiceName] = true
	}
	off := []Icon{}
	for _, d := range loadDisabled() {
		if !live[d.ServiceName] {
			off = append(off, d)
		}
	}
	sort.Slice(icons, func(i, j int) bool { return icons[i].ServiceName < icons[j].ServiceName })
	sort.Slice(off, func(i, j int) bool { return off[i].ServiceName < off[j].ServiceName })
	outJSON(200, map[string]any{
		"data":     append(icons, off...),
		"total":    len(icons) + len(off),
		"enabled":  len(icons),
		"disabled": len(off),
		"is_admin": isAdmin,
	})
}

func apiIconFile(query string) {
	var f string
	for _, kv := range strings.Split(query, "&") {
		if strings.HasPrefix(kv, "f=") {
			f, _ = urlUnescape(kv[2:])
		}
	}
	if f == "" {
		fmt.Println("Status: 404 Not Found")
		fmt.Println()
		return
	}
	base := filepath.Base(f)
	if base == "" || base == "." || base == ".." || strings.ContainsAny(base, `/\`) {
		fmt.Println("Status: 400 Bad Request")
		fmt.Println()
		return
	}
	b, err := os.ReadFile(filepath.Join(imagesDir(), base))
	if err != nil {
		fmt.Println("Status: 404 Not Found")
		fmt.Println()
		return
	}
	ct := "application/octet-stream"
	switch strings.ToLower(filepath.Ext(base)) {
	case ".png":
		ct = "image/png"
	case ".jpg", ".jpeg":
		ct = "image/jpeg"
	case ".svg":
		ct = "image/svg+xml"
	}
	fmt.Println("Content-Type: " + ct)
	fmt.Println("Cache-Control: public, max-age=60")
	fmt.Println()
	os.Stdout.Write(b)
}

func urlUnescape(s string) (string, error) {
	// 简易百分号解码
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+2 < len(s) {
			var v int
			_, err := fmt.Sscanf(s[i+1:i+3], "%02x", &v)
			if err == nil {
				b.WriteByte(byte(v))
				i += 2
				continue
			}
		}
		if s[i] == '+' {
			b.WriteByte(' ')
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String(), nil
}

func apiService(query string) {
	var name string
	for _, kv := range strings.Split(query, "&") {
		if strings.HasPrefix(kv, "name=") {
			name, _ = urlUnescape(kv[5:])
		}
	}
	if name == "" {
		outErr(400, "缺少 name")
		return
	}
	var ic Icon
	err := db.QueryRow(`
		SELECT id, service_name, coalesce(title,''), coalesce(icon,''),
		       coalesce("type",''), coalesce(url,''), is_admin, no_display
		FROM app_service WHERE app_id=$1 AND service_name=$2`, selfAppID, name).
		Scan(&ic.ID, &ic.ServiceName, &ic.Title, &ic.Icon, &ic.Type, &ic.URL,
			&ic.IsAdmin, &ic.NoDisplay)
	if err == sql.ErrNoRows {
		outErr(404, "不存在")
		return
	}
	if err != nil {
		outErr(500, err.Error())
		return
	}
	ic.IconURL = iconURL(ic.Icon)
	outJSON(200, ic)
}

type payload struct {
	ServiceName string `json:"service_name"`
	Title       string `json:"title"`
	Type        string `json:"type"`
	URL         string `json:"url"`
	IsAdmin     bool   `json:"is_admin"`
	IconBase64  string `json:"icon_base64"`
}

func (p *payload) normalize() error {
	p.ServiceName = strings.TrimSpace(p.ServiceName)
	p.Title = strings.TrimSpace(p.Title)
	p.Type = strings.TrimSpace(p.Type)
	p.URL = strings.TrimSpace(p.URL)
	if p.Title == "" {
		return fmt.Errorf("图标名称不能为空")
	}
	if p.Type != "url" && p.Type != "iframe" {
		return fmt.Errorf("打开方式必须是 url 或 iframe")
	}
	if p.URL == "" {
		return fmt.Errorf("访问地址不能为空")
	}
	return nil
}

func apiCreate(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		outErr(400, "请求格式错误")
		return
	}
	if err := p.normalize(); err != nil {
		outErr(400, err.Error())
		return
	}
	if p.ServiceName == "" {
		p.ServiceName = "customIcon." + strconv.FormatInt(time.Now().Unix(), 10)
	}
	if !strings.HasPrefix(p.ServiceName, "customIcon.") && !strings.HasPrefix(p.ServiceName, "iconManager.") {
		p.ServiceName = "customIcon." + p.ServiceName
	}

	var n int
	db.QueryRow(`SELECT count(*) FROM app_service WHERE service_name=$1`, p.ServiceName).Scan(&n)
	if n > 0 {
		outErr(409, "服务名称已存在")
		return
	}

	iconPath := "ui/images/icon_{0}.png"
	if p.IconBase64 != "" {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(p.IconBase64, "data:image/png;base64,"))
		if err == nil {
			if name, e := saveIconSizes(esc(p.ServiceName), raw); e == nil {
				iconPath = "ui/images/" + name
			}
		}
	}

	var newID int64
	err := db.QueryRow(`
		INSERT INTO app_service (app_id, service_name, title, icon, "type", url,
		                         default_url, is_admin, control, full_url, no_display)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,'',false) RETURNING id`,
		selfAppID, p.ServiceName, p.Title, iconPath, p.Type, p.URL, p.URL,
		p.IsAdmin, defaultControl).Scan(&newID)
	if err != nil {
		outErr(500, "新增失败: "+err.Error())
		return
	}
	logf.Printf("新增 %s id=%d", p.ServiceName, newID)
	outJSON(200, map[string]any{"ok": true, "id": newID, "message": "新增成功"})
}

func apiUpdate(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var p payload
	if err := json.Unmarshal(body, &p); err != nil {
		outErr(400, "请求格式错误")
		return
	}
	if p.ServiceName == "" {
		outErr(400, "缺少 service_name")
		return
	}
	if err := p.normalize(); err != nil {
		outErr(400, err.Error())
		return
	}

	var cur string
	err := db.QueryRow(`SELECT coalesce(icon,'') FROM app_service WHERE app_id=$1 AND service_name=$2`,
		selfAppID, p.ServiceName).Scan(&cur)
	if err == sql.ErrNoRows {
		outErr(404, "不存在")
		return
	}
	if err != nil {
		outErr(500, err.Error())
		return
	}

	iconPath := cur
	if p.IconBase64 != "" {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(p.IconBase64, "data:image/png;base64,"))
		if err == nil {
			if name, e := saveIconSizes(esc(p.ServiceName), raw); e == nil {
				iconPath = "ui/images/" + name
			}
		}
	}

	_, err = db.Exec(`UPDATE app_service SET title=$1, icon=$2, "type"=$3, url=$4,
		default_url=$5, is_admin=$6 WHERE app_id=$7 AND service_name=$8`,
		p.Title, iconPath, p.Type, p.URL, p.URL, p.IsAdmin, selfAppID, p.ServiceName)
	if err != nil {
		outErr(500, "更新失败: "+err.Error())
		return
	}
	outJSON(200, map[string]any{"ok": true, "message": "更新成功"})
}

func apiDelete(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var req struct {
		ServiceName string `json:"service_name"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ServiceName == "" {
		outErr(400, "缺少 service_name")
		return
	}
	res, err := db.Exec(`DELETE FROM app_service WHERE app_id=$1 AND service_name=$2`,
		selfAppID, req.ServiceName)
	if err != nil {
		outErr(500, err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		outErr(404, "不存在")
		return
	}
	os.RemoveAll(backupDir(req.ServiceName))
	logf.Printf("永久删除 %s", req.ServiceName)
	outJSON(200, map[string]any{"ok": true, "message": "已删除"})
}

func apiDisable(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var req struct {
		ServiceName string `json:"service_name"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ServiceName == "" {
		outErr(400, "缺少 service_name")
		return
	}

	var rowJSON []byte
	err := db.QueryRow(`SELECT row_to_json(t) FROM (
		SELECT * FROM app_service WHERE app_id=$1 AND service_name=$2) t`,
		selfAppID, req.ServiceName).Scan(&rowJSON)
	if err == sql.ErrNoRows {
		outErr(404, "不存在")
		return
	}
	if err != nil {
		outErr(500, "读取失败: "+err.Error())
		return
	}

	dir := backupDir(req.ServiceName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		outErr(500, "创建备份目录失败")
		return
	}
	rec := BackupRecord{ServiceName: req.ServiceName,
		DisabledAt: time.Now().Format(time.RFC3339), Row: rowJSON}
	blob, _ := json.MarshalIndent(rec, "", "  ")
	recPath := filepath.Join(dir, "record.json")
	if err := os.WriteFile(recPath, blob, 0o644); err != nil {
		outErr(500, "写备份失败")
		return
	}
	check, err := os.ReadFile(recPath)
	if err != nil || len(check) == 0 || !strings.Contains(string(check), `"service_name"`) {
		outErr(500, "备份校验失败，已中止")
		return
	}

	var tmp map[string]any
	if json.Unmarshal(rowJSON, &tmp) == nil {
		if v, ok := tmp["icon"].(string); ok && v != "" {
			if b, e := os.ReadFile(filepath.Join(imagesDir(), filepath.Base(v))); e == nil {
				os.WriteFile(filepath.Join(dir, filepath.Base(v)), b, 0o644)
			}
		}
	}

	res, err := db.Exec(`DELETE FROM app_service WHERE app_id=$1 AND service_name=$2`,
		selfAppID, req.ServiceName)
	if err != nil {
		outErr(500, "删除失败（备份保留）: "+err.Error())
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		outErr(404, "未删除")
		return
	}
	logf.Printf("关闭 %s", req.ServiceName)
	outJSON(200, map[string]any{"ok": true, "message": "已关闭"})
}

func apiEnable(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var req struct {
		ServiceName string `json:"service_name"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.ServiceName == "" {
		outErr(400, "缺少 service_name")
		return
	}

	dir := backupDir(req.ServiceName)
	blob, err := os.ReadFile(filepath.Join(dir, "record.json"))
	if err != nil {
		outErr(404, "找不到备份")
		return
	}
	var rec BackupRecord
	if err := json.Unmarshal(blob, &rec); err != nil || len(rec.Row) == 0 {
		outErr(500, "备份损坏")
		return
	}

	var exist int
	db.QueryRow(`SELECT count(*) FROM app_service WHERE app_id=$1 AND service_name=$2`,
		selfAppID, req.ServiceName).Scan(&exist)
	if exist > 0 {
		outErr(409, "已存在")
		return
	}

	var tmp map[string]any
	if json.Unmarshal(rec.Row, &tmp) == nil {
		if v, ok := tmp["icon"].(string); ok && v != "" {
			base := filepath.Base(v)
			dst := filepath.Join(imagesDir(), base)
			if _, e := os.Stat(dst); os.IsNotExist(e) {
				if b, e2 := os.ReadFile(filepath.Join(dir, base)); e2 == nil {
					os.WriteFile(dst, b, 0o644)
				}
			}
		}
	}

	var newID int64
	err = db.QueryRow(`INSERT INTO app_service
		SELECT * FROM json_populate_record(NULL::app_service, $1::json) RETURNING id`,
		string(rec.Row)).Scan(&newID)
	if err != nil {
		outErr(500, "恢复失败: "+err.Error())
		return
	}
	arch := filepath.Join(backupRoot(), "_restored")
	os.MkdirAll(arch, 0o755)
	os.Rename(dir, filepath.Join(arch, esc(req.ServiceName)+"-"+time.Now().Format("20060102-150405")))
	logf.Printf("开启 %s id=%d", req.ServiceName, newID)
	outJSON(200, map[string]any{"ok": true, "id": newID, "message": "已开启"})
}

func apiUpload(body []byte, isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	var req struct {
		FileName string `json:"file_name"`
		Base64   string `json:"base64"`
	}
	if err := json.Unmarshal(body, &req); err != nil || req.Base64 == "" {
		outErr(400, "缺少图片数据")
		return
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(req.Base64, "data:image/png;base64,"))
	if err != nil {
		outErr(400, "base64 解析失败")
		return
	}
	if len(raw) > 2<<20 {
		outErr(400, "图片不能超过 2MB")
		return
	}
	name := esc(strings.TrimSuffix(req.FileName, filepath.Ext(req.FileName)))
	if name == "" {
		name = fmt.Sprintf("upload_%d", time.Now().Unix())
	}
	fn, err := saveIconSizes(name, raw)
	if err != nil {
		outErr(500, err.Error())
		return
	}
	outJSON(200, map[string]any{"ok": true, "icon_name": fn})
}

func apiReload(isAdmin bool) {
	if !isAdmin {
		outErr(403, "仅管理员可操作")
		return
	}
	reloadMu.Lock()
	defer reloadMu.Unlock()
	logf.Println("重载 trim_app_center ...")
	// CGI 是短命进程，不能同步等 systemd 重启（会被 SIGKILL 打断）
	// 用 systemd-run 脱离当前会话，异步执行
	cmd := exec.Command("systemd-run", "--no-block", "--collect",
		"--unit=iconmgr-reload-"+strconv.FormatInt(time.Now().UnixNano(), 10),
		"systemctl", "restart", serviceUnit)
	out, err := cmd.CombinedOutput()
	if err != nil {
		// 退路：直接 nohup 后台跑
		sh := exec.Command("sh", "-c",
			"nohup systemctl restart "+serviceUnit+" >/dev/null 2>&1 &")
		if e2 := sh.Start(); e2 != nil {
			outErr(500, "重载失败: "+err.Error()+" / "+string(out)+" / "+e2.Error())
			return
		}
		sh.Process.Release()
	}
	logf.Println("已提交重载任务")
	outJSON(200, map[string]any{"ok": true, "message": "已提交重载，约 5 秒后生效"})
}
