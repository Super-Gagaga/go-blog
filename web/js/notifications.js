const API = '/api/v1';
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function show(msg) { $('#status').textContent = msg; $('#status').hidden = !msg; }
function date(v) { return v ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(v)) : ''; }
async function api(path, options = {}) {
  if (!token()) throw new Error('请先登录后查看消息。');
  const res = await fetch(API + path, { ...options, headers: { Authorization: `Bearer ${token()}`, 'Content-Type': 'application/json', ...(options.headers || {}) } });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || '请求失败');
  return body.data;
}
function parseContent(raw) { try { return JSON.parse(raw || '{}'); } catch { return {}; } }
function icon(type) {
  return { comment:'ph-chat-circle', reply:'ph-chats', like:'ph-heart', follow:'ph-user-plus', review_approved:'ph-check-circle', review_rejected:'ph-warning-circle' }[type] || 'ph-bell';
}
function text(n) {
  const c = parseContent(n.content);
  if (n.type === 'review_approved') return ['文章审核通过', `《${c.article_title || '文章'}》已经发布。`];
  if (n.type === 'review_rejected') return ['文章审核未通过', `《${c.article_title || '文章'}》需要修改：${c.reason || ''}`];
  if (n.type === 'follow') return ['新的关注', '有读者关注了你。'];
  if (n.type === 'like') return ['新的喜欢', '你的文章收到了喜欢。'];
  if (n.type === 'reply') return ['新的回复', '你的评论有了新回复。'];
  if (n.type === 'comment') return ['新的评论', '你的文章有了新评论。'];
  return [c.title || '系统消息', c.message || n.content || ''];
}
function render(result) {
  const list = result.list || [];
  if (!list.length) { $('#notice-list').innerHTML = '<div class="empty">暂无通知。</div>'; return; }
  $('#notice-list').innerHTML = list.map(n => {
    const [title, body] = text(n);
    return `<article class="notice-card ${n.is_read ? '' : 'unread'}">
      <div class="notice-icon"><i class="ph ${icon(n.type)}"></i></div>
      <div><div class="notice-title">${esc(title)}</div><div class="notice-body">${esc(body)}</div></div>
      <time class="notice-time">${date(n.created_at)}</time>
    </article>`;
  }).join('');
}
async function load() {
  show('正在加载...');
  try { render(await api('/user/notifications?page=1&page_size=50')); show(''); }
  catch (e) { show(e.message); $('#notice-list').innerHTML = ''; }
}
$('#mark-all').addEventListener('click', async () => {
  try { await api('/user/notifications/read', { method: 'PATCH', body: JSON.stringify({ ids: [] }) }); await load(); }
  catch (e) { show(e.message); }
});
load();
