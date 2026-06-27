const API = '/api/v1';
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
let view = 'pending';
function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function show(msg) { $('#status').textContent = msg; $('#status').hidden = !msg; }
async function api(path, options = {}) {
  if (!token()) { location.href = '/admin-login'; throw new Error('请先登录。'); }
  const res = await fetch(API + path, { ...options, headers: { Authorization: `Bearer ${token()}`, 'Content-Type': 'application/json', ...(options.headers || {}) } });
  if (res.status === 401) { localStorage.removeItem('jwt_token'); localStorage.removeItem('refresh_token'); location.href = '/admin-login'; throw new Error('登录已过期'); }
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || `请求失败 ${res.status}`);
  return body.data;
}
function setTitle(_label, title) { $('#view-title').textContent = title; }
function table(head, rows) {
  return `<div class="table-wrap"><table><thead><tr>${head.map(h => `<th>${h}</th>`).join('')}</tr></thead><tbody>${rows.join('') || `<tr><td colspan="${head.length}"><div class="empty">暂无数据</div></td></tr>`}</tbody></table></div>`;
}
function openModal(title, body) {
  $('#modal-title').textContent = title;
  $('#modal-body').innerHTML = body;
  $('#modal').showModal();
}
function closeModal() { $('#modal').close(); }
function statusChip(s) {
  const text = { draft:'草稿', pending_review:'待审核', published:'已发布', archived:'已归档', active:'正常', banned:'封禁', admin:'管理员', user:'用户' }[s] || s;
  const cls = s === 'published' || s === 'active' ? 'green' : s === 'pending_review' || s === 'admin' ? 'orange' : s === 'banned' ? 'red' : '';
  return `<span class="chip ${cls}">${text}</span>`;
}
async function loadDashboard() {
  setTitle('overview', '仪表盘');
  const d = await api('/admin/dashboard');
  $('#pending-badge').textContent = d.pending_review_count || 0;
  $('#view').innerHTML = `<div class="stats-grid">
    ${[['用户',d.user_count],['管理员',d.admin_count],['文章',d.article_count],['已发布',d.published_count],['待审核',d.pending_review_count],['评论',d.comment_count],['点赞',d.like_count],['收藏',d.favorite_count]].map(x => `<div class="stat-card"><div class="tiny">${x[0]}</div><div class="stat-value">${x[1] || 0}</div></div>`).join('')}
  </div>
  <div class="layout"><section>${table(['最新文章','作者','状态'], (d.recent_articles || []).map(a => `<tr><td>${esc(a.title)}</td><td>${esc(a.user?.username || '')}</td><td>${statusChip(a.status)}</td></tr>`))}</section>
  <aside>${table(['高产作者','文章'], (d.top_authors || []).map(a => `<tr><td>${esc(a.username)}</td><td>${a.article_count}</td></tr>`))}</aside></div>`;
}
async function loadPending() {
  setTitle('review queue', '待审核文章');
  const data = await api('/admin/articles/pending?page=1&page_size=50');
  $('#pending-badge').textContent = data.pagination?.total || 0;
  $('#view').innerHTML = table(['标题','作者','提交时间','操作'], (data.list || []).map(a => `<tr>
    <td><strong>${esc(a.title)}</strong><div class="tiny">${esc(a.summary || '').slice(0, 80)}</div></td>
    <td>${esc(a.user?.username || '')}</td><td>${new Date(a.created_at).toLocaleString('zh-CN')}</td>
    <td class="row-actions"><button class="btn" data-preview="${a.id}">预览</button><button class="btn" data-approve="${a.id}">通过</button><button class="btn danger" data-reject="${a.id}">驳回</button></td>
  </tr>`));
  (data.list || []).forEach(a => {
    document.querySelector(`[data-preview="${a.id}"]`)?.addEventListener('click', () => openModal(a.title, `<div class="preview-content">${a.content_html || `<p>${esc(a.content || '')}</p>`}</div>`));
    document.querySelector(`[data-approve="${a.id}"]`)?.addEventListener('click', () => action(`/admin/articles/${a.id}/approve`, 'POST'));
    document.querySelector(`[data-reject="${a.id}"]`)?.addEventListener('click', () => rejectArticle(a.id));
  });
}
function rejectArticle(id) {
  openModal('驳回文章', `<div class="form-grid"><label class="field"><span>驳回原因</span><textarea id="reject-reason" class="textarea" maxlength="500"></textarea></label><button class="btn danger" id="confirm-reject" type="button">确认驳回</button></div>`);
  $('#confirm-reject').onclick = () => action(`/admin/articles/${id}/reject`, 'POST', { reason: $('#reject-reason').value.trim() }, closeModal);
}
async function loadArticles() {
  setTitle('content', '文章管理');
  $('#view').innerHTML = `<div class="filters"><input class="input" id="kw" placeholder="搜索标题"><select class="select" id="st"><option value="">全部状态</option><option value="pending_review">待审核</option><option value="published">已发布</option><option value="draft">草稿</option><option value="archived">归档</option></select><button class="btn" id="filter">筛选</button></div><div id="table"></div>`;
  const run = async () => {
    const data = await api(`/admin/articles?page=1&page_size=50&keyword=${encodeURIComponent($('#kw').value)}&status=${encodeURIComponent($('#st').value)}`);
    $('#table').innerHTML = table(['标题','作者','状态','阅读','操作'], (data.list || []).map(a => `<tr><td>${esc(a.title)}</td><td>${esc(a.user?.username || '')}</td><td>${statusChip(a.status)}</td><td>${a.view_count || 0}</td><td><button class="btn danger" data-del-article="${a.id}">删除</button></td></tr>`));
    document.querySelectorAll('[data-del-article]').forEach(b => b.onclick = () => confirm('确定删除这篇文章？') && action(`/admin/articles/${b.dataset.delArticle}`, 'DELETE'));
  };
  $('#filter').onclick = run; await run();
}
async function loadComments() {
  setTitle('moderation', '评论管理');
  const data = await api('/admin/comments?page=1&page_size=50');
  $('#view').innerHTML = table(['内容','评论者','文章','时间','操作'], (data.list || []).map(c => `<tr><td>${esc(c.content).slice(0, 90)}</td><td>${esc(c.user?.username || '')}</td><td>${esc(c.article?.title || '')}</td><td>${new Date(c.created_at).toLocaleString('zh-CN')}</td><td><button class="btn danger" data-del-comment="${c.id}">删除</button></td></tr>`));
  document.querySelectorAll('[data-del-comment]').forEach(b => b.onclick = () => confirm('确定删除这条评论？') && action(`/admin/comments/${b.dataset.delComment}`, 'DELETE'));
}
async function loadUsers() {
  setTitle('people', '用户管理');
  const data = await api('/admin/users?page=1&page_size=50');
  $('#view').innerHTML = table(['用户','邮箱','角色','状态','操作'], (data.list || []).map(u => `<tr><td>${esc(u.username)}</td><td>${esc(u.email)}</td><td>${statusChip(u.role)}</td><td>${statusChip(u.status)}</td><td class="row-actions"><button class="btn" data-role="${u.id}" data-next="${u.role === 'admin' ? 'user' : 'admin'}">${u.role === 'admin' ? '取消管理员' : '设为管理员'}</button><button class="btn danger" data-ban="${u.id}" data-next="${u.status === 'banned' ? 'active' : 'banned'}">${u.status === 'banned' ? '解封' : '封禁'}</button></td></tr>`));
  document.querySelectorAll('[data-role]').forEach(b => b.onclick = () => action(`/admin/users/${b.dataset.role}/role`, 'PATCH', { role: b.dataset.next }));
  document.querySelectorAll('[data-ban]').forEach(b => b.onclick = () => action(`/admin/users/${b.dataset.ban}/status`, 'PATCH', { status: b.dataset.next }));
}
async function loadTaxonomies() {
  setTitle('taxonomy', '分类标签');
  const [cats, tags] = await Promise.all([fetch(API + '/categories').then(r => r.json()).then(b => b.data || []), fetch(API + '/tags').then(r => r.json()).then(b => b.data || [])]);
  $('#view').innerHTML = `<div class="layout"><section><div class="actions"><input id="cat-name" class="input" placeholder="新分类"><button id="add-cat" class="btn">新增分类</button></div>${table(['分类','操作'], cats.map(c => `<tr><td>${esc(c.name)}</td><td><button class="btn danger" data-del-cat="${c.id}">删除</button></td></tr>`))}</section><aside><div class="actions"><input id="tag-name" class="input" placeholder="新标签"><button id="add-tag" class="btn">新增标签</button></div>${table(['标签','操作'], tags.map(t => `<tr><td>${esc(t.name)}</td><td><button class="btn danger" data-del-tag="${t.id}">删除</button></td></tr>`))}</aside></div>`;
  $('#add-cat').onclick = () => action('/admin/categories', 'POST', { name: $('#cat-name').value.trim() });
  $('#add-tag').onclick = () => action('/admin/tags', 'POST', { name: $('#tag-name').value.trim() });
  document.querySelectorAll('[data-del-cat]').forEach(b => b.onclick = () => action(`/admin/categories/${b.dataset.delCat}`, 'DELETE'));
  document.querySelectorAll('[data-del-tag]').forEach(b => b.onclick = () => action(`/admin/tags/${b.dataset.delTag}`, 'DELETE'));
}
async function loadAudit() {
  setTitle('audit', '审计日志');
  const data = await api('/admin/audit-logs?page=1&page_size=80');
  $('#view').innerHTML = table(['管理员','动作','目标','详情','IP','时间'], (data.list || []).map(l => `<tr><td>${esc(l.admin?.username || l.admin_id)}</td><td>${esc(l.action)}</td><td>${esc(l.target_type)} #${l.target_id}</td><td>${esc(l.detail || '').slice(0, 120)}</td><td>${esc(l.ip || '')}</td><td>${new Date(l.created_at).toLocaleString('zh-CN')}</td></tr>`));
}
async function action(path, method, body, after) {
  try {
    await api(path, { method, body: body ? JSON.stringify(body) : undefined });
    if (after) after();
    await load();
  } catch (e) { show(e.message); }
}
async function guard() {
  if (!token()) {
    location.href = '/admin-login';
    throw new Error('请先登录');
  }
  const res = await fetch(API + '/user/profile', {
    headers: { Authorization: `Bearer ${token()}` }
  });
  if (res.status === 401) {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
    location.href = '/admin-login';
    throw new Error('登录已过期');
  }
  const body = await res.json().catch(() => ({}));
  if (body.code !== 0 || !body.data) {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
    location.href = '/admin-login';
    throw new Error('无法获取用户信息');
  }
  if (body.data.role !== 'admin') {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
    location.href = '/admin-login?error=' + encodeURIComponent('此账号没有管理员权限');
    throw new Error('当前账号没有管理员权限。');
  }
}
async function load() {
  show('正在加载...');
  try {
    const loaders = { pending: loadPending, dashboard: loadDashboard, articles: loadArticles, comments: loadComments, users: loadUsers, taxonomies: loadTaxonomies, audit: loadAudit };
    await loaders[view]();
    show('');
  } catch (e) { show(e.message); $('#view').innerHTML = ''; }
}
document.querySelectorAll('.admin-nav-item').forEach(btn => btn.onclick = () => {
  view = btn.dataset.view;
  document.querySelectorAll('.admin-nav-item').forEach(x => x.classList.toggle('active', x === btn));
  load();
});
$('#refresh').onclick = load;
$('#logout').onclick = () => { localStorage.removeItem('jwt_token'); localStorage.removeItem('refresh_token'); location.href = '/'; };
(async function init() { try { await guard(); await load(); } catch (e) { show(e.message); } })();
