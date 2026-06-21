const API = '/api/v1';
let tab = 'following';
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function avatar(u = {}) { return u.avatar || `https://picsum.photos/seed/${encodeURIComponent(u.username || u.id || 'user')}/100/100`; }
function show(msg) { $('#status').textContent = msg; $('#status').hidden = !msg; }
async function api(path, options = {}) {
  if (!token()) throw new Error('请先登录后查看关注关系。');
  const res = await fetch(API + path, { ...options, headers: { Authorization: `Bearer ${token()}`, ...(options.headers || {}) } });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || '请求失败');
  return body.data;
}
function pickUser(item) { return item.followee || item.follower || item.user || item; }
function render(result) {
  const list = result.list || [];
  if (!list.length) { $('#follow-list').innerHTML = `<div class="empty">${tab === 'following' ? '你还没有关注作者。' : '暂时还没有人关注你。'}</div>`; return; }
  $('#follow-list').innerHTML = list.map(item => {
    const u = pickUser(item);
    return `<article class="person-card">
      <div class="person-head"><img class="avatar" src="${esc(avatar(u))}" alt=""><div><div class="person-name">${esc(u.username || '用户')}</div><div class="tiny">${esc(u.email || '')}</div></div></div>
      <p class="person-bio">${esc(u.bio || '这个作者还没有填写简介。')}</p>
      <div class="person-stats"><span>${u.article_count || 0} articles</span><span>${u.follower_count || 0} followers</span></div>
    </article>`;
  }).join('');
}
async function load() {
  show('正在加载...');
  try { render(await api(`/user/${tab}?page=1&page_size=50`)); show(''); }
  catch (e) { show(e.message); $('#follow-list').innerHTML = ''; }
}
document.querySelectorAll('.follow-tab').forEach(btn => btn.addEventListener('click', () => {
  tab = btn.dataset.tab;
  document.querySelectorAll('.follow-tab').forEach(x => x.classList.toggle('active', x === btn));
  load();
}));
load();
