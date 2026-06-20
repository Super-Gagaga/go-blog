const API = '/api/v1';
const qs = new URLSearchParams(location.search);
const articleID = qs.get('id');
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
const statusEl = $('#status');
const articleEl = $('#article');
const commentsPanel = $('#comments-panel');
const commentList = $('#comment-list');

function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function showStatus(msg) { statusEl.textContent = msg; statusEl.hidden = !msg; }
function avatar(u = {}) { return u.avatar || `https://picsum.photos/seed/${encodeURIComponent(u.username || 'journal')}/100/100`; }
function cover(a = {}) { return a.cover_image || `https://picsum.photos/seed/${encodeURIComponent(a.slug || a.title || a.id || 'article')}/1200/640`; }
function date(v) { return v ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium' }).format(new Date(v)) : ''; }
async function request(path, options = {}) {
  const headers = { ...(options.headers || {}) };
  if (token()) headers.Authorization = `Bearer ${token()}`;
  const res = await fetch(API + path, { ...options, headers });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || `请求失败 ${res.status}`);
  return body.data;
}
function renderArticle(a) {
  document.title = `${a.title} - Journal`;
  articleEl.innerHTML = `
    <p class="label">${esc(a.category?.name || 'article')}</p>
    <h1 class="page-title">${esc(a.title)}</h1>
    <div class="article-meta">
      <img class="avatar" src="${esc(avatar(a.user))}" alt="">
      <span>${esc(a.user?.username || '匿名作者')}</span>
      <span>${date(a.created_at)}</span>
      <span>${a.view_count || 0} 阅读</span>
      <span>${a.like_count || 0} 喜欢</span>
      <span>${a.favorite_count || 0} 收藏</span>
    </div>
    <img class="article-cover" src="${esc(cover(a))}" alt="${esc(a.title)}">
    <div class="article-tools">
      <button class="btn" data-action="like"><i class="ph ph-heart"></i>喜欢</button>
      <button class="btn" data-action="favorite"><i class="ph ph-bookmark-simple"></i>收藏</button>
      <a class="btn" href="/editor?id=${a.id}"><i class="ph ph-pencil-simple"></i>编辑</a>
    </div>
    <div class="article-content">${a.content_html || `<p>${esc(a.content || a.summary || '')}</p>`}</div>
  `;
  articleEl.hidden = false;
  articleEl.querySelector('[data-action="like"]').onclick = () => toggle(`/articles/${a.id}/like`, '已记录喜欢');
  articleEl.querySelector('[data-action="favorite"]').onclick = () => toggle(`/articles/${a.id}/favorite`, '已更新收藏');
}
async function toggle(path, okText) {
  if (!token()) { showStatus('请先登录后再操作。'); return; }
  try { await request(path, { method: 'POST' }); showStatus(okText); setTimeout(() => showStatus(''), 1500); }
  catch (e) { showStatus(e.message); }
}
function renderComments(nodes = []) {
  $('#comment-count').textContent = `${nodes.length} 条评论`;
  if (!nodes.length) { commentList.innerHTML = '<div class="empty">还没有评论，写下第一条讨论。</div>'; return; }
  const nodeHTML = n => `
    <div class="comment-item">
      <div class="comment-top"><img class="avatar" src="${esc(avatar(n.user))}" alt=""><strong>${esc(n.user?.username || '已删除用户')}</strong><span class="tiny">${date(n.created_at)}</span></div>
      <div class="comment-content">${esc(n.content || '')}</div>
      ${(n.replies || []).length ? `<div class="replies">${n.replies.map(nodeHTML).join('')}</div>` : ''}
    </div>`;
  commentList.innerHTML = nodes.map(nodeHTML).join('');
}
async function loadComments() {
  const data = await request(`/articles/${articleID}/comments`);
  renderComments(data || []);
}
$('#comment-form').addEventListener('submit', async e => {
  e.preventDefault();
  if (!token()) { $('#comment-hint').textContent = '请先登录后再评论。'; return; }
  const content = $('#comment-input').value.trim();
  if (!content) return;
  try {
    await request(`/articles/${articleID}/comments`, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ content }) });
    $('#comment-input').value = '';
    $('#comment-hint').textContent = '评论已发布';
    await loadComments();
  } catch (err) { $('#comment-hint').textContent = err.message; }
});
(async function init() {
  if (!articleID) { showStatus('缺少文章 ID。'); return; }
  try {
    const article = await request(`/articles/${articleID}`);
    renderArticle(article);
    commentsPanel.hidden = false;
    showStatus('');
    await loadComments();
  } catch (err) { showStatus(err.message); }
})();
