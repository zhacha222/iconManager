# iconManager —— 飞牛 fnOS 扩展图标管理（带开关）

> 给飞牛 fnOS 的「扩展图标管理」加上**单个图标启用/禁用开关**。
> UI 复刻官方 `customIcon` 扩展，CGI 模式运行，不占端口、无常驻进程。

[![fnOS](https://img.shields.io/badge/fnOS-0.9.27%2B-blue)](https://developer.fnnas.com/docs/guide)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.21%2B-00ADD8)](https://go.dev)

## 为什么做这个

飞牛官方的「扩展图标管理」（`customIcon`）可以增删自定义桌面图标，但**不能单独禁用某个图标**。
想临时隐藏一个图标，只能删除它；删除后它的配置和图标都没了，想恢复得重新填一遍。

本项目在完整保留官方功能的基础上，增加了一个开关：

- **关闭** = 自动备份该图标（整行数据 + 图标文件）后从数据库移除 → 桌面图标消失
- **开启** = 从备份精确还原（**保留原 id、原配置**）→ 桌面图标回来

开关状态会列在列表里（灰色显示），随时可以点开恢复。想彻底删掉再用「删除」按钮。

## 功能

| 功能 | 说明 |
|---|---|
| 图标列表 | 表格展示，支持按服务名称/图标名称搜索、分页 |
| **启用开关** | 关闭 = 备份 + 移除；开启 = 从备份恢复（保留原 id） |
| 新增图标 | 完整表单：服务名称 / 显示名称 / 上传图片 / 访问权限 / URL 类型 / 打开方式 / 端口 / 路径 / 自定义 URL |
| 编辑图标 | 同表单，服务名称不可改 |
| 永久删除 | 移除记录 + 清除备份（与开关区分：开关可恢复，删除不可） |
| 上传图标 | 自动生成 64×64 和 256×256 两个尺寸 |
| 重载服务 | 重启 `trim_app_center`，让桌面立即刷新 |
| 主题切换 | 亮色 / 暗色，记忆在 localStorage |

## 安装

### 前置要求

- 飞牛 fnOS **0.9.27 或更高**
- x86_64 架构
- 具备管理员权限

### 方式一：安装 fpk 包

从 [Releases](../../releases) 下载最新的 `iconManager.fpk`，然后在应用中心「手动安装」选择该文件。

或用命令行：

```bash
sudo appcenter-cli install-fpk iconManager.fpk
sudo appcenter-cli start iconManager
```

安装后桌面会出现「图标管理·开关」图标，点开即可使用。

### 方式二：从源码构建

需要 Go 1.21+ 和 [fnpack](https://developer.fnnas.com/docs/cli/fnpack/)：

```bash
# 1. 编译 CGI 二进制（-buildvcs=false 保证同一份源码产出相同二进制）
cd src
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -trimpath -buildvcs=false -ldflags="-s -w" -o ../app/ui/index.cgi .

# 2. 打包
cd ..
fnpack build          # 产出 iconManager.fpk
```

## 使用

### 访问地址

```
https://<你的NAS域名>/cgi/ThirdParty/iconManager/index.cgi/
```

也可以直接点桌面上的「图标管理·开关」图标（在飞牛窗口内打开）。

### 开关怎么生效

改完图标后**桌面不会自动刷新** —— 飞牛的应用中心（`trim_app_center`）有内存缓存。

所以流程是：

```
拨开关  →  点右上角「重载服务」  →  等 3~8 秒  →  刷新桌面
```

「重载服务」会重启 `trim_app_center.service`，这是让桌面读新数据的唯一方式。

## 工作原理

### 运行模型：CGI

与官方 `customIcon` 一致，采用 CGI 按需启动：

| | iconManager |
|---|---|
| 运行方式 | 飞牛每次请求启动一次进程，处理完退出 |
| 占用端口 | **不占** |
| 空闲资源 | **0**（无常驻进程） |
| 是否 Docker | 否，原生 fpk 应用 |

### 数据结构

图标存在 `appcenter` 数据库的 `app_service` 表，本应用通过 **unix socket + peer 认证**直连：

```
飞牛桌面 (iframe)
    ↓
trim_http_cgi  →  ui/index.cgi (Go)
                       ↓ unix socket (peer auth)
                  PostgreSQL appcenter.app_service
                       ↓
                  ui/images/*.png
```

### 开关的备份格式

关闭时在 `<数据目录>/backups/<service_name>/` 下存：

```
record.json          # 整行 JSON + 禁用时间
<icon>.png           # 图标文件原件
```

数据目录默认由安装位置推导（`<卷>/@appcenter/iconManager` → `<卷>/@appdata/iconManager`），
可用 `ICONMGR_DATA_DIR` 覆盖。

开启时用 PostgreSQL 的 `json_populate_record` 精确还原**全部字段**（含原 `id`）。

## 重要注意事项

### 1. 需要 `run-as: root`

飞牛的 `pg_hba.conf` 配置为 `local all all md5`，只有 `root` 能通过 `map=trim_root`
peer 认证直连数据库。这与官方 `customIcon` 的做法一致。

本应用只做四件事：读表、删行、插行、读写图标文件。
所有 SQL 使用参数化查询，图标文件名拒绝路径穿越。

### 2. 卸载应用会删除它的图标记录

`appcenter-cli uninstall` 会清掉该 `app_id` 下所有 `app_service` 行。

**卸载前请先备份**：

```bash
sudo -u postgres psql -d appcenter \
  -c "COPY (SELECT * FROM app_service WHERE app_id=(SELECT id FROM app WHERE app_name='iconManager')) TO STDOUT WITH (FORMAT csv, HEADER)" \
  > iconManager-backup.csv
```

或者用「停用」代替「卸载」。

### 3. 升级请替换文件，不要卸载重装

卸载重装会让 `app.id` 变化。本应用已做兼容（运行时按 `app_name` 查 id），
但重装仍会**丢失所有图标记录**。

推荐升级方式：

```bash
sudo appcenter-cli stop iconManager
# 用新版本替换 /vol2/@appcenter/iconManager/ui/index.cgi
sudo cp new-index.cgi /vol2/@appcenter/iconManager/ui/index.cgi
sudo chmod +x /vol2/@appcenter/iconManager/ui/index.cgi
sudo appcenter-cli start iconManager
```

### 4. 目录推导

CGI 由 `trim_http_cgi` 拉起，实测**不注入** `TRIM_APPDEST` / `TRIM_PKGVAR`，
只提供标准 CGI 变量。因此路径从**自身可执行文件位置**推导：

| 目录 | 推导规则 | 覆盖变量 |
|---|---|---|
| 安装目录 | `<APPDIR>/ui/index.cgi` → `<APPDIR>` | `ICONMGR_APPDIR` |
| 图标目录 | `<APPDIR>/ui/images` | `ICONMGR_IMAGES_DIR` |
| 数据目录（备份/日志） | `<卷>/@appcenter/<app>` → `<卷>/@appdata/<app>` | `ICONMGR_DATA_DIR` |

这样不依赖卷号（`vol1`/`vol2`/`vol3` 均可），也不需要修改系统配置。

## 与官方 customIcon 的关系

本项目修改的是**同一张表**（`app_service`）的**同一批数据**。

如果你想把官方扩展的图标迁移到本项目接管，见 `migrate.sh`：

```bash
bash migrate.sh check      # 检查环境与数据量
bash migrate.sh backup     # 备份 app_service 和 app 表
bash migrate.sh run        # 执行迁移（改 app_id + 拷图标文件）
bash migrate.sh verify     # 验证
bash migrate.sh rollback   # 回滚
```

迁移后官方 `customIcon` 可以停用（建议先停用而非卸载，确认无问题后再卸）。

## 项目结构

```
iconManager/
├── src/
│   ├── main.go           # 后端：DB 访问、备份/恢复、增删改、图片缩放、重载
│   ├── ui.go             # 前端（内嵌单文件 HTML）
│   ├── go.mod
│   └── go.sum
├── app/
│   └── ui/
│       ├── config        # 桌面入口声明
│       └── images/       # 入口图标
├── config/
│   ├── privilege         # run-as: root
│   └── resource          # {}
├── cmd/
│   └── main              # 生命周期脚本（CGI 模式无需常驻进程）
├── manifest              # 应用元信息
├── ICON.PNG              # 64×64 包图标
├── ICON_256.PNG          # 256×256 包图标
├── migrate.sh            # 从官方扩展迁移的脚本
└── README.md
```

## UI 复刻说明

样式取自官方 `customIcon` 的 scoped CSS，配色用 Element Plus 官方色板：

```css
.icon-list-container { padding:10px 10px 0 }
.search-container    { background:#f5f7fa; padding:20px 20px 0; border-radius:4px;
                       display:flex; justify-content:space-between }
.pagination-container{ margin-top:20px; text-align:right }
.copyright-container { margin-top:20px; padding:20px 0 0;
                       border-top:1px solid #e4e7ed; color:#909399; font-size:14px }
.form-container      { background:#fff; padding:30px; border-radius:4px;
                       box-shadow:0 2px 12px #0000001a }
```

相对官方版本，**新增的只有「启用状态」一列**和顶部「重载服务」按钮。

## 致谢

- 原「扩展图标管理」由 **MR_XIAOBO / 晓勃网络 (XIAOBONET)** 开发
- 应用规范参考 [飞牛开发者文档](https://developer.fnnas.com/docs/guide)

## 许可

[MIT](LICENSE)。第三方组件与参考项目的声明见 [NOTICE.md](NOTICE.md)。

本项目与飞牛（fnOS）官方无隶属关系，为社区独立开发。
