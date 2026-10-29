// 前端页面（内嵌进二进制）
//
// UI 复刻官方「扩展图标管理」（customIcon 2.3.9-Beta），并增加单个图标开关。
// 官方样式取自 /vol2/@appcenter/customIcon/dist/index.html 的 scoped CSS：
//   .icon-list-container{padding:10px 10px 0}
//   .search-container{background:#f5f7fa;padding:20px 20px 0;border-radius:4px;display:flex;justify-content:space-between}
//   .pagination-container{margin-top:20px;text-align:right}
//   .copyright-container{margin-top:20px;padding:20px 0 0;border-top:1px solid #e4e7ed;color:#909399;font-size:14px}
//   .icon-form-container{padding:20px} / .form-container{background:#fff;padding:30px;border-radius:4px;box-shadow:0 2px 12px #0000001a}
package main

const indexHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<base href="/cgi/ThirdParty/iconManager/index.cgi/">

<title>扩展图标管理</title>
<style>
:root{
  --primary:#409eff; --primary-l3:#79bbff; --primary-l9:#ecf5ff;
  --success:#67c23a; --warning:#e6a23c; --danger:#f56c6c; --info:#909399;
  --text-0:#303133; --text-1:#606266; --text-2:#909399; --text-3:#c0c4cc;
  --border:#dcdfe6; --border-l:#e4e7ed; --border-l2:#ebeef5;
  --bg:#fff; --bg-page:#f0f2f5; --bg-fill:#f5f7fa;
  --radius:4px;
}
html.dark{
  --primary:#409eff; --primary-l3:#3375b9; --primary-l9:#18222c;
  --text-0:#e5eaf3; --text-1:#cfd3dc; --text-2:#a3a6ad; --text-3:#8d9095;
  --border:#4c4d4f; --border-l:#414243; --border-l2:#363637;
  --bg:#1d1e1f; --bg-page:#141414; --bg-fill:#262727;
}
*{box-sizing:border-box}
body{
  margin:0;padding:10px;background:var(--bg-page);color:var(--text-0);
  font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC",
       "Hiragino Sans GB","Microsoft YaHei",Arial,sans-serif;
}
body,.modal,.row,.el-input__inner{transition:background-color .2s,color .2s}

/* ---------- 通用控件（模仿 Element Plus） ---------- */
button{font:inherit;cursor:pointer;outline:none}
.el-btn{
  display:inline-flex;align-items:center;justify-content:center;gap:5px;
  padding:8px 15px;font-size:14px;border-radius:var(--radius);line-height:1;
  border:1px solid var(--border);background:var(--bg);color:var(--text-1);
  transition:.15s;white-space:nowrap;
}
.el-btn:hover{border-color:var(--primary-l3);color:var(--primary);background:var(--primary-l9)}
.el-btn--primary{background:var(--primary);border-color:var(--primary);color:#fff}
.el-btn--primary:hover{background:var(--primary-l3);border-color:var(--primary-l3);color:#fff}
.el-btn--warning{background:var(--warning);border-color:var(--warning);color:#fff}
.el-btn--warning:hover{background:#ebb563;border-color:#ebb563;color:#fff}
.el-btn--danger{color:var(--danger);border-color:#fbc4c4;background:#fef0f0}
.el-btn--danger:hover{background:var(--danger);border-color:var(--danger);color:#fff}
.el-btn--mini{padding:5px 9px;font-size:12px;border-radius:3px}
.el-btn.is-disabled{opacity:.5;cursor:not-allowed}
.el-btn.is-loading{opacity:.65;cursor:wait}
.btn-group{display:inline-flex}
.btn-group .el-btn{border-radius:0;margin-left:-1px}
.btn-group .el-btn:first-child{border-radius:var(--radius) 0 0 var(--radius);margin-left:0}
.btn-group .el-btn:last-child{border-radius:0 var(--radius) var(--radius) 0}

.el-input{
  display:inline-block;position:relative;
}
.el-input__inner{
  width:100%;padding:0 11px;height:32px;line-height:32px;font:inherit;
  color:var(--text-1);background:var(--bg);border:1px solid var(--border);
  border-radius:var(--radius);outline:none;transition:.15s;
}
.el-input__inner:hover{border-color:var(--text-3)}
.el-input__inner:focus{border-color:var(--primary)}
.el-input__inner::placeholder{color:var(--text-3)}

/* ---------- 列表页（官方 .icon-list-container） ---------- */
.header{display:flex;justify-content:space-between;align-items:center;margin-bottom:16px}
.header h1{margin:0;color:var(--text-0);font-size:20px;font-weight:600}

.icon-list-container{padding:10px 10px 0}

.search-container{
  background:var(--bg-fill);padding:20px 20px 0;border-radius:var(--radius);
  display:flex;justify-content:space-between;gap:16px;flex-wrap:wrap;
}
.search-left{display:flex;gap:16px;align-items:flex-start;flex-wrap:wrap;flex:1}
.search-right{display:flex;gap:10px;align-items:flex-start;padding-bottom:18px}
.form-item{display:flex;align-items:center;gap:8px;margin-bottom:18px}
.form-item > label{
  font-size:14px;color:var(--text-1);white-space:nowrap;
  width:auto;text-align:right;padding-right:2px;
}

/* 表格（模仿 el-table） */
.table-wrap{background:var(--bg);border:1px solid var(--border-l2);border-radius:var(--radius);
            overflow:hidden;margin-top:16px}
table{width:100%;border-collapse:separate;border-spacing:0;font-size:14px}
thead th{
  background:var(--bg-fill);color:var(--text-1);font-weight:600;text-align:left;
  padding:10px 12px;border-bottom:1px solid var(--border-l2);font-size:13px;
  white-space:nowrap;
}
tbody td{padding:10px 12px;border-bottom:1px solid var(--border-l2);vertical-align:middle}
tbody tr:last-child td{border-bottom:none}
tbody tr:hover td{background:var(--primary-l9)}
tbody tr.off td{opacity:.55}

.tag{
  display:inline-block;padding:0 8px;height:22px;line-height:20px;font-size:12px;
  border-radius:3px;border:1px solid;white-space:nowrap;
}
.tag--primary{background:var(--primary-l9);border-color:#b3d8ff;color:var(--primary)}
.tag--warning{background:#fdf6ec;border-color:#f5dab1;color:var(--warning)}
.tag--success{background:#f0f9eb;border-color:#c2e7b0;color:var(--success)}
.tag--info{background:var(--bg-fill);border-color:var(--border-l);color:var(--text-2)}
html.dark .tag--primary{background:#18222c;border-color:#1a3250}
html.dark .tag--warning{background:#292218;border-color:#4d3a1a}
html.dark .tag--success{background:#1c2518;border-color:#274723}

.thumb{width:44px;height:44px;border-radius:6px;object-fit:cover;display:block;
       background:var(--bg-fill);border:1px solid var(--border-l2)}
.thumb-ph{width:44px;height:44px;border-radius:6px;display:grid;place-items:center;
          background:var(--bg-fill);color:var(--text-3);border:1px solid var(--border-l2);font-size:16px}
.mono{font-family:ui-monospace,Consolas,monospace;font-size:12.5px;color:var(--text-2);
      word-break:break-all}
.title-cell{font-weight:500;color:var(--text-0);word-break:break-all}

/* 分页 */
.pagination-container{margin-top:20px;text-align:right;display:flex;
  justify-content:flex-end;align-items:center;gap:10px;color:var(--text-1);font-size:14px}
.pagination-container .el-btn{padding:5px 11px;font-size:13px}

/* 版权（官方原样保留） */
.copyright-container{
  margin-top:20px;padding:20px 0 0;border-top:1px solid var(--border-l);
  color:var(--text-2);font-size:14px;
}
.copyright-content{display:flex;align-items:center;justify-content:center;gap:20px;
                   max-width:820px;margin:0 auto;flex-wrap:wrap}
.copyright-content p{margin:2px 0}
.copyright-content a{color:var(--primary);text-decoration:none}
.copyright-content a:hover{text-decoration:underline}

/* ---------- 表单页（官方 .icon-form-container） ---------- */
.icon-form-container{padding:20px}
.form-container{
  background:var(--bg);padding:30px;border-radius:var(--radius);
  box-shadow:0 2px 12px rgba(0,0,0,.1);
}
.form-container .el-form-item{display:flex;align-items:flex-start;margin-bottom:22px;gap:12px}
.form-container .el-form-item > label{
  width:110px;flex-shrink:0;text-align:right;padding-top:7px;
  font-size:14px;color:var(--text-1);line-height:1.4;
}
.form-container .el-form-item > label .req{color:var(--danger);margin-right:2px}
.form-container .el-form-item .field{flex:1;min-width:0;max-width:600px}

.radios{display:inline-flex;flex-wrap:wrap;gap:0}
.radios input{display:none}
.radios label{
  padding:6px 14px;border:1px solid var(--border);background:var(--bg);
  color:var(--text-1);font-size:14px;cursor:pointer;margin:0 0 0 -1px;
  display:inline-flex;align-items:center;gap:5px;line-height:20px;
}
.radios label:first-child{border-radius:var(--radius) 0 0 var(--radius);margin-left:0}
.radios label:last-child{border-radius:0 var(--radius) var(--radius) 0}
.radios input:checked + span{color:var(--primary);font-weight:500}
.radios label:has(input:checked){
  border-color:var(--primary);background:var(--primary-l9);z-index:1;position:relative;
}
.hint{font-size:12px;color:var(--text-2);line-height:1.6;margin-top:6px}

/* 上传 */
.image-upload-container{width:100%}
.upload-area{
  border:1px dashed var(--border);border-radius:6px;background:var(--bg-fill);
  padding:24px;text-align:center;cursor:pointer;transition:.15s;color:var(--text-2);
}
.upload-area:hover{border-color:var(--primary);color:var(--primary)}
.upload-area.has{padding:14px}
.upload-area img{width:88px;height:88px;border-radius:8px;object-fit:cover;
                 display:block;margin:0 auto 8px;border:1px solid var(--border-l2)}

select.el-input__inner{cursor:pointer}

/* 弹窗（确认框） */
.mask{position:fixed;inset:0;background:rgba(0,0,0,.5);display:none;
      align-items:center;justify-content:center;z-index:2000;padding:20px}
.mask.show{display:flex}
.msgbox{background:var(--bg);border-radius:8px;box-shadow:0 12px 32px rgba(0,0,0,.2);
        width:100%;max-width:430px;overflow:hidden}
.msgbox .mb-head{padding:18px 20px 6px;font-size:16px;font-weight:500;color:var(--text-0)}
.msgbox .mb-body{padding:0 20px 20px;color:var(--text-1);font-size:14px;line-height:1.7}
.msgbox .mb-body b{color:var(--danger)}
.msgbox .mb-foot{padding:0 20px 18px;display:flex;justify-content:flex-end;gap:10px}

/* 开关（el-switch 风格） */
.sw{position:relative;width:40px;height:20px;display:inline-block;vertical-align:middle}
.sw input{opacity:0;width:0;height:0;position:absolute}
.sl{position:absolute;inset:0;background:#dcdfe6;border-radius:999px;cursor:pointer;transition:.2s}
.sl:before{content:"";position:absolute;height:16px;width:16px;left:2px;top:2px;background:#fff;
  border-radius:50%;transition:.2s;box-shadow:0 1px 2px rgba(0,0,0,.2)}
.sw input:checked + .sl{background:#67c23a}
.sw input:checked + .sl:before{transform:translateX(20px)}
.sw input:disabled + .sl{cursor:not-allowed;opacity:.5}
html.dark .sl{background:#4c4d4f}
/* toast */
#toast{position:fixed;left:50%;top:20px;transform:translateX(-50%) translateY(-80px);
  background:var(--bg);color:var(--text-0);padding:11px 18px;border-radius:6px;
  border:1px solid var(--border-l);box-shadow:0 4px 16px rgba(0,0,0,.15);
  font-size:14px;opacity:0;transition:.25s;pointer-events:none;z-index:3000;
  max-width:90vw;text-align:center}
#toast.show{opacity:1;transform:translateX(-50%) translateY(0)}
#toast.ok{border-color:#c2e7b0;background:#f0f9eb;color:var(--success)}
#toast.err{border-color:#fbc4c4;background:#fef0f0;color:var(--danger)}
html.dark #toast.ok{background:#1c2518;border-color:#274723}
html.dark #toast.err{background:#2b1d1d;border-color:#472626}

.loading{padding:60px;text-align:center;color:var(--text-2)}
@media(max-width:860px){
  .search-container{flex-direction:column}
  .search-right{padding-bottom:0}
  .form-container{padding:18px}
  .form-container .el-form-item{flex-direction:column;align-items:stretch;gap:6px}
  .form-container .el-form-item > label{width:auto;text-align:left;padding-top:0}
  .hide-sm{display:none}
}
</style>
</head>
<body>

<!-- ================= 列表页 ================= -->
<div class="icon-list-container" id="pageList">
  <div class="header">
    <h1>扩展图标管理</h1>
    <div style="display:flex;gap:10px">
      <button class="el-btn" id="btnTheme" title="切换主题">◐ 主题</button>
      <button class="el-btn" id="btnReload" title="改库后桌面图标需重载服务才刷新">↻ 重载服务</button>
    </div>
  </div>

  <div class="search-container">
    <div class="search-left">
      <div class="form-item">
        <label>服务名称</label>
        <div class="el-input" style="width:140px">
          <input class="el-input__inner" id="qName" placeholder="请输入服务名称">
        </div>
      </div>
      <div class="form-item">
        <label>图标名称</label>
        <div class="el-input" style="width:140px">
          <input class="el-input__inner" id="qTitle" placeholder="请输入图标名称">
        </div>
      </div>
      <div class="form-item">
        <button class="el-btn el-btn--primary" id="btnSearch">搜索</button>
        <button class="el-btn" id="btnReset">重置</button>
      </div>
    </div>
    <div class="search-right">
      <button class="el-btn el-btn--primary" id="btnNew">新增图标</button>
    </div>
  </div>

  <div class="table-wrap">
    <table>
      <thead>
        <tr>
          <th style="width:64px">图标预览</th>
          <th style="width:200px">图标名称</th>
          <th style="width:150px">服务名称</th>
          <th style="width:88px">打开方式</th>
          <th class="hide-sm">访问地址</th>
          <th style="width:96px">启用状态</th>
          <th style="width:130px">操作</th>
        </tr>
      </thead>
      <tbody id="tbody"></tbody>
    </table>
  </div>

  <div class="pagination-container">
    <span id="pageInfo"></span>
    <div class="btn-group">
      <button class="el-btn" id="btnPrev">上一页</button>
      <button class="el-btn" id="btnNext">下一页</button>
    </div>
  </div>

  <div class="copyright-container">
    <div class="copyright-content">
      <div>
        <p>基于 MR_XIAOBO 的「扩展图标管理」重构，增加单图标开关</p>
        <p>原项目 Copyright © 2021-Present 晓勃网络(XIAOBONET). All Rights Reserved.</p>
      </div>
    </div>
  </div>
</div>

<!-- ================= 表单页 ================= -->
<div class="icon-form-container" id="pageForm" style="display:none">
  <div class="header">
    <h1 id="formTitle">新增图标</h1>
    <button class="el-btn" id="btnBack">返回列表</button>
  </div>

  <div class="form-container">
    <div class="el-form-item">
      <label><span class="req">*</span>服务名称</label>
      <div class="field">
        <div class="el-input">
          <input class="el-input__inner" id="fName" placeholder="服务名称仅支持字母、数字、下划线组合">
        </div>
        <div class="hint">会自动加 <code>customIcon.</code> 前缀；编辑时不可修改</div>
      </div>
    </div>

    <div class="el-form-item">
      <label><span class="req">*</span>图标显示名称</label>
      <div class="field">
        <div class="el-input">
          <input class="el-input__inner" id="fTitle" placeholder="请输入图标显示名称,大于14字符,会超出边界">
        </div>
      </div>
    </div>

    <div class="el-form-item">
      <label>图标上传</label>
      <div class="field">
        <div class="image-upload-container">
          <div class="upload-area" id="upBox">选择图片</div>
        </div>
        <div class="hint">建议 256×256 PNG/JPG，不超过 1024 KB；留空使用默认图标</div>
      </div>
    </div>

    <div class="el-form-item">
      <label>访问权限</label>
      <div class="field">
        <div class="radios">
          <label><input type="radio" name="fAdmin" value="f"><span>设备内所有用户</span></label>
          <label><input type="radio" name="fAdmin" value="t" checked><span>仅管理员可见</span></label>
        </div>
      </div>
    </div>

    <div class="el-form-item">
      <label>URL类型</label>
      <div class="field">
        <div class="radios">
          <label><input type="radio" name="fUrlType" value="2" checked><span>外部链接</span></label>
          <label><input type="radio" name="fUrlType" value="1"><span>NAS内部</span></label>
        </div>
      </div>
    </div>

    <div class="el-form-item">
      <label>打开方式</label>
      <div class="field">
        <select class="el-input__inner" id="fType">
          <option value="url">浏览器新标签页</option>
          <option value="iframe">窗口模式(仅支持NAS内http协议及应用中心cgi页面)</option>
        </select>
      </div>
    </div>

    <div class="el-form-item">
      <label>访问端口</label>
      <div class="field">
        <div class="el-input">
          <input class="el-input__inner" id="fPort" placeholder="不填写即为NAS默认端口, 或输入应用端口号">
        </div>
      </div>
    </div>

    <div class="el-form-item">
      <label><span class="req">*</span>访问路径</label>
      <div class="field">
        <div class="el-input">
          <input class="el-input__inner" id="fPath" placeholder="内部链接输入相对访问路径，外部链接输入完整URL地址">
        </div>
        <div class="hint">路径必须以 / 开头或 http/https 开头</div>
      </div>
    </div>

    <div class="el-form-item">
      <label>自定义URL</label>
      <div class="field">
        <div class="el-input">
          <input class="el-input__inner" id="fFullUrl" placeholder="反向代理或公网访问时，可设置自定义URL">
        </div>
        <div class="hint">反向代理用户：可设置自定义反向代理地址，在非局域网模式下，将优先使用该地址进行访问。</div>
      </div>
    </div>

    <div class="el-form-item">
      <label></label>
      <div class="field">
        <button class="el-btn el-btn--primary" id="btnSave">保存</button>
        <button class="el-btn" id="btnCancel" style="margin-left:10px">取消</button>
      </div>
    </div>
  </div>
</div>

<!-- 确认框 -->
<div class="mask" id="mask">
  <div class="msgbox">
    <div class="mb-head" id="mbTitle">提示</div>
    <div class="mb-body" id="mbBody"></div>
    <div class="mb-foot">
      <button class="el-btn" id="mbCancel">取消</button>
      <button class="el-btn el-btn--primary" id="mbOk">确定</button>
    </div>
  </div>
</div>

<div id="toast"></div>

<script>
const $ = s => document.querySelector(s);
const $$ = s => Array.from(document.querySelectorAll(s));

let ICONS = [], IS_ADMIN = false, STAT = {enabled:0, disabled:0};
let EDITING = null, UPLOADED = null;
let PAGE = 1, PAGE_SIZE = 10;
let QNAME = '', QTITLE = '';

/* ---------- 主题 ---------- */
(function(){
  const saved = localStorage.getItem('iconmgr-theme');
  if(saved === 'dark') document.documentElement.classList.add('dark');
})();
$('#btnTheme').addEventListener('click', () => {
  const d = document.documentElement.classList.toggle('dark');
  localStorage.setItem('iconmgr-theme', d ? 'dark' : 'light');
});

/* ---------- 工具 ---------- */
function rel(u){ return String(u || "").replace(/^\/+/, ""); }
function esc(s){
  return String(s == null ? '' : s).replace(/[&<>"']/g,
    c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
}
function toast(msg, kind){
  const t = $('#toast');
  t.textContent = msg;
  t.className = 'show' + (kind ? ' ' + kind : '');
  clearTimeout(t._tm);
  t._tm = setTimeout(() => t.className = '', 2800);
}
async function api(path, opts){
  const u = new URL(String(path).replace(/^\/+/, ""), document.baseURI).href;
  const r = await fetch(u, opts);
  const ct = r.headers.get('content-type') || '';
  const body = ct.includes('json') ? await r.json() : await r.text();
  if(!r.ok) throw new Error((body && body.error) || ('HTTP ' + r.status));
  return body;
}
function post(path, data){
  return api(path, {
    method:'POST',
    headers:{'Content-Type':'application/json'},
    body:JSON.stringify(data || {})
  });
}

/* ---------- 确认框 ---------- */
let mbResolve = null;
function confirmBox(title, html, okText){
  return new Promise(res => {
    $('#mbTitle').textContent = title || '提示';
    $('#mbBody').innerHTML = html || '';
    $('#mbOk').textContent = okText || '确定';
    $('#mask').classList.add('show');
    mbResolve = res;
  });
}
function closeMb(v){
  $('#mask').classList.remove('show');
  if(mbResolve){ mbResolve(v); mbResolve = null; }
}
$('#mbOk').addEventListener('click', () => closeMb(true));
$('#mbCancel').addEventListener('click', () => closeMb(false));
$('#mask').addEventListener('click', e => { if(e.target === $('#mask')) closeMb(false); });

/* ---------- 列表 ---------- */
async function load(){
  try{
    const res = await api('/api/icons');
    ICONS = res.data || [];
    IS_ADMIN = !!res.is_admin;
    STAT.enabled = res.enabled ?? 0;
    STAT.disabled = res.disabled ?? 0;
    render();
  }catch(e){
    $('#tbody').innerHTML = '<tr><td colspan="7"><div class="loading">加载失败：'
      + esc(e.message) + '</div></td></tr>';
  }
}

function filtered(){
  return ICONS.filter(it => {
    if(QNAME && !it.service_name.toLowerCase().includes(QNAME)) return false;
    if(QTITLE && !(it.title || '').toLowerCase().includes(QTITLE)) return false;
    return true;
  });
}

function render(){
  const all = filtered();
  const totalPages = Math.max(1, Math.ceil(all.length / PAGE_SIZE));
  if(PAGE > totalPages) PAGE = totalPages;
  const start = (PAGE - 1) * PAGE_SIZE;
  const pageData = all.slice(start, start + PAGE_SIZE);

  if(!ICONS.length){
    $('#tbody').innerHTML = '<tr><td colspan="7"><div class="loading">'
      + '还没有图标，点右上角「新增图标」开始</div></td></tr>';
  }else if(!pageData.length){
    $('#tbody').innerHTML = '<tr><td colspan="7"><div class="loading">没有匹配的记录</div></td></tr>';
  }else{
    let h = '';
    for(const it of pageData){
      const on = !it.disabled;
      const img = rel(it.icon_url)
        ? '<img class="thumb" src="' + esc(rel(it.icon_url)) + '" alt="" loading="lazy" '
          + 'onerror="this.outerHTML=&quot;<div class=thumb-ph>▦</div>&quot;">'
        : '<div class="thumb-ph">▦</div>';
      const typeTag = it.type === 'iframe'
        ? '<span class="tag tag--warning">窗口</span>'
        : '<span class="tag tag--success">新标签页</span>';

      h += '<tr class="' + (on ? '' : 'off') + '">'
        +  '<td>' + img + '</td>'
        +  '<td class="title-cell">' + esc(it.title || it.service_name) + '</td>'
        +  '<td><span class="tag tag--primary">'
        +     esc(it.service_name.replace('customIcon.', '')) + '</span></td>'
        +  '<td>' + typeTag + '</td>'
        +  '<td class="hide-sm"><div class="mono">' + esc(it.url) + '</div></td>'
        +  '<td><label class="sw"><input type="checkbox" ' + (on ? 'checked' : '')
        +     (IS_ADMIN ? '' : ' disabled')
        +     ' data-toggle="' + esc(it.service_name) + '"><span class="sl"></span></label></td>'
        +  '<td><div class="btn-group">'
        +     '<button class="el-btn el-btn--mini el-btn--primary" data-edit="'
        +       esc(it.service_name) + '">编辑</button>'
        +     '<button class="el-btn el-btn--mini el-btn--danger" data-del="'
        +       esc(it.service_name) + '">删除</button>'
        +  '</div></td>'
        + '</tr>';
    }
    $('#tbody').innerHTML = h;
  }

  $('#pageInfo').textContent = '共 ' + all.length + ' 条 · 启用 ' + STAT.enabled
    + ' · 关闭 ' + STAT.disabled + ' · 第 ' + PAGE + '/' + totalPages + ' 页';
  $('#btnPrev').classList.toggle('is-disabled', PAGE <= 1);
  $('#btnNext').classList.toggle('is-disabled', PAGE >= totalPages);

  $$('[data-toggle]').forEach(cb => cb.addEventListener('change', onToggle));
  $$('[data-edit]').forEach(b => b.addEventListener('click', () => openEdit(b.dataset.edit)));
  $$('[data-del]').forEach(b => b.addEventListener('click', () => onDelete(b.dataset.del)));
}



/* ---------- 开关 ---------- */
async function onToggle(ev){
  const cb = ev.target, name = cb.dataset.toggle, wantOn = cb.checked;
  cb.disabled = true;
  try{
    await post(wantOn ? '/api/enable' : '/api/disable', {service_name:name});
    toast((wantOn ? '已启用' : '已禁用') + '：' + name.replace('customIcon.', ''), 'ok');
    await load();
    setTimeout(() => {
      toast('桌面图标需点右上角「重载服务」才刷新');
    }, 1200);
  }catch(e){
    toast((wantOn ? '启用' : '禁用') + '失败：' + e.message, 'err');
    cb.checked = !wantOn;
  }finally{ cb.disabled = false; }
}

async function onDelete(name){
  const ok = await confirmBox('警告信息',
    '确定要<b>永久删除</b>「' + esc(name.replace('customIcon.','')) + '」吗？<br><br>'
    + '此操作会同时清除它的备份，<b>不可恢复</b>。<br>'
    + '如果只想临时隐藏，请使用左侧的启用开关。', '确定删除');
  if(!ok) return;
  try{
    await post('/api/delete', {service_name:name});
    toast('已删除', 'ok');
    await load();
  }catch(e){ toast('删除失败：' + e.message, 'err'); }
}

/* ---------- 表单 ---------- */
function showList(){ $('#pageForm').style.display = 'none'; $('#pageList').style.display = ''; load(); }
function showForm(){ $('#pageList').style.display = 'none'; $('#pageForm').style.display = ''; }

function resetUpBox(url){
  const box = $('#upBox');
  UPLOADED = null;
  if(url){
    box.className = 'upload-area has';
    box.innerHTML = '<img src="' + esc(url) + '" alt="">点击更换图片';
  }else{
    box.className = 'upload-area';
    box.textContent = '选择图片';
  }
}

function openNew(){
  EDITING = null;
  $('#formTitle').textContent = '新增图标';
  $('#fName').value = '';
  $('#fName').disabled = false;
  $('#fTitle').value = '';
  $('#fPort').value = '';
  $('#fPath').value = '';
  $('#fFullUrl').value = '';
  $('#fType').value = 'url';
  document.querySelector('input[name=fAdmin][value=t]').checked = true;
  document.querySelector('input[name=fUrlType][value="2"]').checked = true;
  resetUpBox(null);
  showForm();
  $('#fName').focus();
}

async function openEdit(name){
  try{
    const it = await api('/api/service?name=' + encodeURIComponent(name));
    EDITING = name;
    $('#formTitle').textContent = '编辑图标';
    $('#fName').value = it.service_name.replace('customIcon.', '');
    $('#fName').disabled = true;
    $('#fTitle').value = it.title || '';
    $('#fType').value = it.type || 'url';

    // 从 url 里拆端口与路径
    let port = '', path = it.url || '';
    const m = path.match(/^https?:\/\/[^/:]+:(\d+)(\/.*)?$/);
    if(m){ port = m[1]; path = m[2] || '/'; }
    $('#fPort').value = port;
    $('#fPath').value = path;
    $('#fFullUrl').value = '';

    document.querySelector('input[name=fAdmin][value="' + (it.is_admin ? 't' : 'f') + '"]').checked = true;
    document.querySelector('input[name=fUrlType][value="' + (path.startsWith('http') ? '2' : '1') + '"]').checked = true;
    resetUpBox(rel(it.icon_url));
    showForm();
  }catch(e){ toast('加载失败：' + e.message, 'err'); }
}

$('#upBox').addEventListener('click', () => {
  const inp = document.createElement('input');
  inp.type = 'file';
  inp.accept = 'image/png,image/jpeg';
  inp.onchange = () => {
    const f = inp.files[0];
    if(!f) return;
    if(f.size > 1024*1024){ toast('图片不能超过 1024 KB', 'err'); return; }
    const rd = new FileReader();
    rd.onload = () => {
      UPLOADED = rd.result;
      const box = $('#upBox');
      box.className = 'upload-area has';
      box.innerHTML = '<img src="' + rd.result + '" alt="">点击更换图片';
    };
    rd.readAsDataURL(f);
  };
  inp.click();
});

/* 组装地址 */
function buildUrl(){
  const urlType = document.querySelector('input[name=fUrlType]:checked').value;
  const port = $('#fPort').value.trim();
  let path = $('#fPath').value.trim();
  if(!path) return '';

  // 外部链接：必须 http(s) 开头
  if(/^https?:\/\//i.test(path)){
    return path;
  }
  // 内网/IP 形式：补协议
  if(/^\d{1,3}(\.\d{1,3}){3}/.test(path) || path.startsWith('localhost')){
    return 'http://' + path + (port && !path.includes(':') ? ':' + port : '');
  }
  // 相对路径
  if(port) return 'http://${host}:' + port + (path.startsWith('/') ? path : '/' + path);
  return path.startsWith('/') ? path : '/' + path;
}

$('#btnSave').addEventListener('click', async () => {
  const url = buildUrl();
  const payload = {
    service_name: $('#fName').value.trim(),
    title: $('#fTitle').value.trim(),
    type: $('#fType').value,
    url: url,
    is_admin: document.querySelector('input[name=fAdmin]:checked').value === 't'
  };
  if(UPLOADED) payload.icon_base64 = UPLOADED;

  if(!payload.title){ toast('请填写图标显示名称', 'err'); return; }
  if(!url){ toast('请填写访问路径', 'err'); return; }
  if(!EDITING && !payload.service_name){ toast('请填写服务名称', 'err'); return; }
  if(payload.type === 'iframe' && !url.startsWith('/')){
    toast('窗口模式需要在访问路径中设置相对路径', 'err'); return;
  }

  const btn = $('#btnSave');
  btn.disabled = true; btn.classList.add('is-loading');
  try{
    if(EDITING){
      await post('/api/update', payload);
      toast('更新成功', 'ok');
    }else{
      await post('/api/create', payload);
      toast('新增成功', 'ok');
    }
    showList();
    setTimeout(() => toast('桌面图标需点右上角「重载服务」才刷新'), 1200);
  }catch(e){
    toast('保存失败：' + e.message, 'err');
  }finally{ btn.disabled = false; btn.classList.remove('is-loading'); }
});

$('#btnCancel').addEventListener('click', showList);
$('#btnBack').addEventListener('click', showList);

/* ---------- 搜索 / 分页 ---------- */
$('#btnSearch').addEventListener('click', () => {
  QNAME = $('#qName').value.trim().toLowerCase();
  QTITLE = $('#qTitle').value.trim().toLowerCase();
  PAGE = 1;
  render();
});
$('#btnReset').addEventListener('click', () => {
  $('#qName').value = ''; $('#qTitle').value = '';
  QNAME = ''; QTITLE = ''; PAGE = 1;
  render();
});
$('#qName').addEventListener('keydown', e => { if(e.key === 'Enter') $('#btnSearch').click(); });
$('#qTitle').addEventListener('keydown', e => { if(e.key === 'Enter') $('#btnSearch').click(); });

$('#btnPrev').addEventListener('click', () => { if(PAGE > 1){ PAGE--; render(); } });
$('#btnNext').addEventListener('click', () => {
  const tp = Math.max(1, Math.ceil(filtered().length / PAGE_SIZE));
  if(PAGE < tp){ PAGE++; render(); }
});

/* ---------- 重载服务 ---------- */
$('#btnReload').addEventListener('click', async () => {
  const ok = await confirmBox('警告信息',
    '确定要重启<b>飞牛应用中心</b>吗？<br><br>'
    + '新增图标后如需立即在桌面看到，必须重载服务。<br>'
    + '过程中约 3-8 秒，桌面可能短暂不可用。', '确认并继续');
  if(!ok) return;

  const b = $('#btnReload');
  b.disabled = true; b.classList.add('is-loading');
  const old = b.textContent;
  b.textContent = '↻ 重载中…';
  try{
    await post('/api/reload');
    toast('服务已重载，请刷新桌面查看', 'ok');
  }catch(e){
    toast('重载失败：' + e.message, 'err');
  }finally{
    b.disabled = false; b.classList.remove('is-loading'); b.textContent = old;
  }
});

$('#btnNew').addEventListener('click', openNew);

load();
</script>
</body>
</html>
`
