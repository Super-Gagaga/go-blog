const API = '/api/v1';
let page = 1;
const pageSize = 10;
const token = () => localStorage.getItem('jwt_token');
const $ = s => document.querySelector(s);
function esc(v = '') { return String(v).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function avatar(u = {}) { return u.avatar || `https://picsum.photos/seed/${encodeURIComponent(u.username || u.id || 'user')}/100/100`; }
function coverImg(a) { return a.cover_image || ''; }
function show(msg) { $('#status').textContent = msg; $('#status').hidden = !msg; }

function timeAgo(dateStr) {
  const now = Date.now();
  const then = new Date(dateStr).getTime();
  const diff = Math.floor((now - then) / 1000);
  if (diff < 60) return '刚刚';
  if (diff < 3600) return Math.floor(diff / 60) + ' 分钟前';
  if (diff < 86400) return Math.floor(diff / 3600) + ' 小时前';
  if (diff < 2592000) return Math.floor(diff / 86400) + ' 天前';
  return new Date(dateStr).toLocaleDateString('zh-CN');
}

async function api(path, options = {}) {
  if (!token()) throw new Error('请先登录后查看收藏。');
  const res = await fetch(API + path, { ...options, headers: { Authorization: `Bearer ${token()}`, ...(options.headers || {}) } });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || '请求失败');
  return body.data;
}

function render(result) {
  const list = result.list || [];
  if (!list.length) {
    $('#favorites-list').innerHTML = '<div class="empty">你还没有收藏文章。<br>在文章详情页点击收藏按钮即可添加到这里。</div>';
    $('#pagination').innerHTML = '';
    return;
  }

  $('#favorites-list').innerHTML = list.map(item => {
    const a = item.article || {};
    const u = a.user || {};
    const cat = a.category || null;
    const tags = a.tags || [];
    const thumb = coverImg(a);

    return `<article class="article-item" data-slug="${esc(a.slug || '')}">
      <div class="article-body">
        <div class="article-title">${esc(a.title || '无标题')}</div>
        <div class="article-excerpt">${esc(a.summary || '')}</div>
        <div class="article-footer">
          <img class="author-avatar" src="${esc(avatar(u))}" alt="">
          <span>${esc(u.username || '用户')}</span>
          <span class="dot"></span>
          <span class="article-stat"><i class="ph ph-eye"></i>${a.view_count || 0}</span>
          <span class="article-stat"><i class="ph ph-heart"></i>${a.like_count || 0}</span>
          <span class="article-stat"><i class="ph ph-chat-circle"></i>${a.comment_count || 0}</span>
          ${cat ? `<span class="inline-tag">${esc(cat.name)}</span>` : ''}
          ${tags.slice(0, 2).map(t => `<span class="inline-tag">${esc(t.name)}</span>`).join('')}
          <span class="article-favored-at"><i class="ph ph-bookmark-simple"></i>${timeAgo(item.created_at)}</span>
        </div>
      </div>
      ${thumb ? `<img class="article-thumb" src="${esc(thumb)}" alt="" loading="lazy">` : ''}
    </article>`;
  }).join('');

  renderPagination(result.pagination);
}

function renderPagination(p) {
  if (!p || p.total_pages <= 1) { $('#pagination').innerHTML = ''; return; }

  let html = '';
  html += `<button ${page <= 1 ? 'disabled' : ''} data-page="${page - 1}"><i class="ph ph-caret-left"></i></button>`;

  const maxButtons = 5;
  let start = Math.max(1, page - Math.floor(maxButtons / 2));
  let end = Math.min(p.total_pages, start + maxButtons - 1);
  if (end - start + 1 < maxButtons) start = Math.max(1, end - maxButtons + 1);

  for (let i = start; i <= end; i++) {
    html += `<button class="${i === page ? 'active' : ''}" data-page="${i}">${i}</button>`;
  }

  html += `<button ${page >= p.total_pages ? 'disabled' : ''} data-page="${page + 1}"><i class="ph ph-caret-right"></i></button>`;
  html += `<span class="page-info">共 ${p.total} 篇</span>`;
  $('#pagination').innerHTML = html;
}

async function load(p = 1) {
  page = p;
  show('正在加载...');
  try {
    const data = await api(`/user/favorites?page=${page}&page_size=${pageSize}`);
    render(data);
    show('');
  } catch (e) {
    show(e.message);
    $('#favorites-list').innerHTML = '';
    $('#pagination').innerHTML = '';
  }
}

$('#favorites-list').addEventListener('click', event => {
  const item = event.target.closest('.article-item');
  if (!item) return;
  const slug = item.dataset.slug;
  if (slug) window.location.href = `/article-detail?slug=${encodeURIComponent(slug)}`;
});

$('#pagination').addEventListener('click', event => {
  const btn = event.target.closest('button');
  if (!btn || btn.disabled) return;
  const targetPage = parseInt(btn.dataset.page, 10);
  if (targetPage && targetPage !== page) {
    load(targetPage);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  }
});

load();
