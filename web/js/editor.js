const API = '/api/v1';
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
const params = new URLSearchParams(location.search);
let editingID = params.get('id');
function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function show(msg) { $('#status').textContent = msg; $('#status').hidden = !msg; }
function showUploadStatus(msg, isError = false) {
  const el = $('#cover-upload-status');
  if (!el) return;
  el.textContent = msg || '';
  el.classList.toggle('error', Boolean(isError));
}
async function api(path, options = {}) {
  if (!token()) throw new Error('请先登录后写作。');
  return apiRequest(token(), path, options);
}

async function apiRequest(tok, path, options = {}) {
  const headers = { Authorization: `Bearer ${tok}`, ...(options.headers || {}) };
  if (options.body && !(options.body instanceof FormData)) headers['Content-Type'] = 'application/json';
  const res = await fetch(API + path, { ...options, headers });
  const body = await res.json().catch(() => ({}));
  if (res.status === 401) {
    const newToken = await tryRefreshToken();
    if (newToken) return apiRequest(newToken, path, options);
  }
  if (!res.ok || body.code !== 0) throw new Error(body.message || body.error || `请求失败 ${res.status}`);
  return body.data;
}

async function tryRefreshToken() {
  const refresh = localStorage.getItem('refresh_token');
  if (!refresh) { show('登录已过期，请返回首页重新登录。'); return null; }
  try {
    const res = await fetch(API + '/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: refresh })
    });
    const body = await res.json().catch(() => ({}));
    if (!res.ok || body.code !== 0) throw new Error();
    localStorage.setItem('jwt_token', body.data.access_token);
    localStorage.setItem('refresh_token', body.data.refresh_token);
    return body.data.access_token;
  } catch {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
    show('登录已过期，请返回首页重新登录。');
    return null;
  }
}
async function loadOptions() {
  const [categories, tags] = await Promise.all([
    fetch(API + '/categories').then(r => r.json()).then(b => b.data || []).catch(() => []),
    fetch(API + '/tags').then(r => r.json()).then(b => b.data || []).catch(() => [])
  ]);
  $('#category').innerHTML = '<option value="">不选择分类</option>' + categories.map(c => `<option value="${c.id}">${esc(c.name)}</option>`).join('');
  $('#tags').innerHTML = tags.map(t => `<option value="${t.id}">${esc(t.name)}</option>`).join('');
}
function payload(status) {
  const tagIDs = Array.from($('#tags').selectedOptions).map(o => Number(o.value));
  const category = $('#category').value;
  return {
    title: $('#title').value.trim(),
    content: $('#content').value,
    category_id: category ? Number(category) : null,
    tag_ids: tagIDs,
    cover_image: $('#cover').value.trim(),
    status
  };
}
function setCover(url) {
  $('#cover').value = url || '';
  const preview = $('#cover-preview');
  if (!preview) return;
  if (url) {
    preview.classList.remove('empty');
    preview.innerHTML = `<img src="${esc(url)}" alt="封面预览">`;
  } else {
    preview.classList.add('empty');
    preview.textContent = '暂未选择封面';
  }
}
function isImage(file) {
  return file && (file.type.startsWith('image/') || /\.(jpe?g|png|webp|gif)$/i.test(file.name || ''));
}
function firstImage(files) {
  return Array.from(files || []).find(isImage);
}
async function walkEntry(entry) {
  if (!entry) return [];
  if (entry.isFile) {
    return new Promise(resolve => entry.file(file => resolve([file]), () => resolve([])));
  }
  if (!entry.isDirectory) return [];
  const reader = entry.createReader();
  const entries = await new Promise(resolve => reader.readEntries(resolve, () => resolve([])));
  const nested = await Promise.all(entries.map(walkEntry));
  return nested.flat();
}
async function filesFromDrop(event) {
  const items = Array.from(event.dataTransfer?.items || []);
  const entries = items.map(item => item.webkitGetAsEntry?.()).filter(Boolean);
  if (entries.length) {
    const nested = await Promise.all(entries.map(walkEntry));
    return nested.flat();
  }
  return Array.from(event.dataTransfer?.files || []);
}
function openUploadModal() {
  showUploadStatus('');
  $('#cover-upload-modal').classList.add('visible');
  $('#cover-upload-modal').setAttribute('aria-hidden', 'false');
}
function closeUploadModal() {
  $('#cover-upload-modal').classList.remove('visible');
  $('#cover-upload-modal').setAttribute('aria-hidden', 'true');
  document.querySelector('#open-cover-upload')?.focus();
}
async function uploadCover(file) {
  if (!file) {
    showUploadStatus('请选择一张图片。', true);
    return;
  }
  if (!isImage(file)) {
    showUploadStatus('只能上传 jpg、png、webp 或 gif 图片。', true);
    return;
  }
  if (file.size > 5 * 1024 * 1024) {
    showUploadStatus('图片不能超过 5MB。', true);
    return;
  }
  showUploadStatus('正在上传封面...');
  try {
    const form = new FormData();
    form.append('image', file);
    const data = await api('/articles/upload', { method: 'POST', body: form });
    setCover(data.url || '');
    showUploadStatus('封面上传成功。');
    setTimeout(closeUploadModal, 450);
  } catch (e) {
    showUploadStatus(e.message, true);
  }
}
async function save(status) {
  const data = payload(status);
  if (!data.title || !data.content.trim()) { show('标题和正文都需要填写。'); return; }
  show(status === 'published' ? '正在提交审核...' : '正在保存草稿...');
  try {
    const path = editingID ? `/articles/${editingID}` : '/articles';
    const method = editingID ? 'PUT' : 'POST';
    const article = await api(path, { method, body: JSON.stringify(data) });
    editingID = article.id;
    history.replaceState(null, '', `/editor?id=${editingID}`);
    show(status === 'published' ? '已提交审核。' : '草稿已保存。');
    loadMine();
  } catch (e) { show(e.message); }
}
async function loadArticle() {
  if (!editingID) return;
  show('正在载入文章...');
  try {
    const a = await api(`/articles/${editingID}`);
    $('#title').value = a.title || '';
    $('#content').value = a.content || '';
    setCover(a.cover_image || '');
    if (a.category_id) $('#category').value = String(a.category_id);
    const tagSet = new Set((a.tags || []).map(t => String(t.id)));
    Array.from($('#tags').options).forEach(o => { o.selected = tagSet.has(o.value); });
    show(a.review_comment ? `驳回原因：${a.review_comment}` : '');
  } catch (e) { show(e.message); }
}
function statusChip(s) {
  const map = { draft:'草稿', pending_review:'待审核', published:'已发布', archived:'已归档' };
  const color = s === 'published' ? 'green' : s === 'pending_review' ? 'orange' : '';
  return `<span class="chip ${color}">${map[s] || s}</span>`;
}
async function loadMine() {
  try {
    const st = $('#article-status').value;
    const data = await api(`/user/articles?page=1&page_size=30${st ? `&status=${encodeURIComponent(st)}` : ''}`);
    const list = data.list || [];
    $('#my-articles').innerHTML = list.length ? list.map(a => `<article class="my-article" data-id="${a.id}">
      <div class="my-article-title">${esc(a.title)}</div>
      <div class="my-article-meta">${statusChip(a.status)}<span class="tiny">${new Date(a.updated_at || a.created_at).toLocaleDateString('zh-CN')}</span></div>
      ${a.review_comment ? `<div class="tiny">驳回：${esc(a.review_comment)}</div>` : ''}
    </article>`).join('') : '<div class="empty">还没有文章。</div>';
    document.querySelectorAll('.my-article').forEach(el => el.onclick = () => location.href = `/editor?id=${el.dataset.id}`);
  } catch (e) { $('#my-articles').innerHTML = `<div class="empty">${esc(e.message)}</div>`; }
}
$('#save-draft').onclick = () => save('draft');
$('#submit-review').onclick = () => save('published');
$('#refresh-mine').onclick = loadMine;
$('#article-status').onchange = loadMine;
$('#open-cover-upload').onclick = openUploadModal;
$('#cover-upload-close').onclick = closeUploadModal;
$('#cover-upload-modal').addEventListener('click', event => {
  if (event.target === $('#cover-upload-modal')) closeUploadModal();
});
$('#choose-cover-file').onclick = () => {
  $('#cover-file-input').value = '';
  $('#cover-file-input').click();
};
$('#choose-cover-folder').onclick = () => {
  $('#cover-folder-input').value = '';
  $('#cover-folder-input').click();
};
$('#cover-file-input').addEventListener('change', event => uploadCover(firstImage(event.target.files)));
$('#cover-folder-input').addEventListener('change', event => uploadCover(firstImage(event.target.files)));
['dragenter', 'dragover'].forEach(type => {
  $('#cover-dropzone').addEventListener(type, event => {
    event.preventDefault();
    $('#cover-dropzone').classList.add('dragover');
  });
});
['dragleave', 'drop'].forEach(type => {
  $('#cover-dropzone').addEventListener(type, event => {
    event.preventDefault();
    $('#cover-dropzone').classList.remove('dragover');
  });
});
$('#cover-dropzone').addEventListener('drop', async event => {
  uploadCover(firstImage(await filesFromDrop(event)));
});
document.addEventListener('keydown', event => {
  if (event.key === 'Escape' && $('#cover-upload-modal')?.classList.contains('visible')) closeUploadModal();
});
(async function init() {
  if (!token()) { show('请先从首页登录，再进入编辑器。'); return; }
  await loadOptions();
  await loadArticle();
  await loadMine();
})();
