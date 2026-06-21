const API = '/api/v1';
const token = () => localStorage.getItem('jwt_token');
const $ = selector => document.querySelector(selector);

let currentProfile = null;

function esc(value = '') {
  return String(value).replace(/[&<>"']/g, ch => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
  }[ch]));
}

function avatarFor(user = {}) {
  if (user.avatar) return user.avatar;
  return `https://picsum.photos/seed/${encodeURIComponent(user.username || 'journal')}/160/160`;
}

function date(value) {
  return value ? new Intl.DateTimeFormat('zh-CN', { dateStyle: 'medium' }).format(new Date(value)) : '';
}

function statusText(status) {
  return { draft: '草稿', pending_review: '待审核', published: '已发布', archived: '已归档' }[status] || status || '';
}

function showStatus(message) {
  const el = $('#status');
  el.textContent = message;
  el.hidden = !message;
}

async function api(path, params = {}, options = {}) {
  if (!token()) {
    location.href = '/';
    return null;
  }
  const url = new URL(API + path, location.origin);
  Object.entries(params).forEach(([key, value]) => {
    if (value !== '' && value !== null && value !== undefined) url.searchParams.set(key, value);
  });
  const res = await fetch(url, {
    ...options,
    headers: { Authorization: `Bearer ${token()}`, ...(options.headers || {}) }
  });
  if (res.status === 401) {
    localStorage.removeItem('jwt_token');
    localStorage.removeItem('refresh_token');
    location.href = '/';
    return null;
  }
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || `请求失败 ${res.status}`);
  return body.data;
}

function renderProfile(profile) {
  currentProfile = profile;
  $('#profile-avatar').src = avatarFor(profile);
  $('#profile-name').textContent = profile.username || '我的主页';
  $('#profile-bio').textContent = profile.bio || '还没有填写简介。';
  $('#article-count').textContent = profile.article_count || 0;
  $('#follower-count').textContent = profile.follower_count || 0;
  $('#following-count').textContent = profile.following_count || 0;
  $('#profile-hero').hidden = false;
  $('#stat-grid').hidden = false;
}

function renderArticles(result) {
  const list = (result && result.list) || [];
  $('#my-articles').innerHTML = list.length ? list.map(article => `
    <a class="profile-item" href="/editor?id=${article.id}">
      <div class="profile-item-title">${esc(article.title || '未命名文章')}</div>
      <div class="profile-item-meta">
        <span>${esc(statusText(article.status))}</span>
        <span>${date(article.updated_at || article.created_at)}</span>
        <span>${article.view_count || 0} 阅读</span>
        <span>${article.like_count || 0} 喜欢</span>
      </div>
    </a>
  `).join('') : '<div class="empty">还没有文章，去写第一篇吧。</div>';
}

function renderFavorites(result) {
  const list = ((result && result.list) || []).map(item => item.article || item).filter(Boolean);
  $('#my-favorites').innerHTML = list.length ? list.slice(0, 8).map(article => `
    <a class="profile-item" href="/article-detail?id=${article.id}">
      <div class="profile-item-title">${esc(article.title || '未命名文章')}</div>
      <div class="profile-item-meta">
        <span>${date(article.created_at)}</span>
        <span>${article.favorite_count || 0} 收藏</span>
      </div>
    </a>
  `).join('') : '<div class="empty">还没有收藏文章。</div>';
}

// --- Edit profile ---

function openEditPanel() {
  if (!currentProfile) return;
  $('#edit-username').value = currentProfile.username || '';
  $('#edit-bio').value = currentProfile.bio || '';
  $('#edit-error').textContent = '';
  $('#profile-edit-panel').hidden = false;
  $('#profile-edit-panel').scrollIntoView({ behavior: 'smooth', block: 'center' });
  setTimeout(() => $('#edit-username').focus(), 150);
}

function closeEditPanel() {
  $('#profile-edit-panel').hidden = true;
  $('#edit-error').textContent = '';
}

async function handleEditSubmit(event) {
  event.preventDefault();
  const submitBtn = $('#profile-edit-form').querySelector('.btn.primary');
  const errorEl = $('#edit-error');
  submitBtn.disabled = true;
  errorEl.textContent = '';

  const username = $('#edit-username').value.trim();
  const bio = $('#edit-bio').value.trim();

  try {
    const body = {};
    if (username && username !== currentProfile.username) body.username = username;
    if (bio !== (currentProfile.bio || '')) body.bio = bio;

    if (!body.username && body.bio === undefined && (!('bio' in body) || Object.keys(body).length === 0)) {
      closeEditPanel();
      return;
    }

    const res = await fetch(`${API}/user/profile`, {
      method: 'PUT',
      headers: { Authorization: `Bearer ${token()}`, 'Content-Type': 'application/json' },
      body: JSON.stringify(body)
    });
    if (res.status === 401) {
      localStorage.removeItem('jwt_token');
      localStorage.removeItem('refresh_token');
      location.href = '/';
      return;
    }
    const data = await res.json();
    if (!res.ok || data.code !== 0) throw new Error(data.message || '保存失败');

    currentProfile = data.data;
    renderProfile(currentProfile);
    closeEditPanel();
  } catch (error) {
    errorEl.textContent = error.message;
  } finally {
    submitBtn.disabled = false;
  }
}

async function handleAvatarChange(event) {
  const file = event.target.files && event.target.files[0];
  if (!file) return;

  const formData = new FormData();
  formData.append('avatar', file);

  showStatus('正在上传头像...');
  try {
    const res = await fetch(`${API}/user/avatar`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token()}` },
      body: formData
    });
    if (res.status === 401) {
      localStorage.removeItem('jwt_token');
      localStorage.removeItem('refresh_token');
      location.href = '/';
      return;
    }
    const data = await res.json();
    if (!res.ok || data.code !== 0) throw new Error(data.message || '上传失败');

    // Refresh the avatar in-place
    const url = data.data.avatar;
    $('#profile-avatar').src = url + '?t=' + Date.now();
    if (currentProfile) currentProfile.avatar = url;
    showStatus('');
  } catch (error) {
    showStatus(error.message);
  } finally {
    $('#avatar-input').value = '';
  }
}

// --- Init ---

(async function init() {
  try {
    showStatus('正在加载我的主页...');
    const [profile, articles, favorites] = await Promise.all([
      api('/user/profile'),
      api('/user/articles', { page: 1, page_size: 10 }),
      api('/user/favorites', { page: 1, page_size: 8 })
    ]);
    if (!profile) return;
    renderProfile(profile);
    renderArticles(articles);
    renderFavorites(favorites);
    $('#profile-content').hidden = false;
    showStatus('');

    // Bind edit events
    $('#edit-profile-btn').addEventListener('click', openEditPanel);
    $('#cancel-edit-btn').addEventListener('click', closeEditPanel);
    $('#cancel-edit').addEventListener('click', closeEditPanel);
    $('#profile-edit-form').addEventListener('submit', handleEditSubmit);
    $('#avatar-edit-btn').addEventListener('click', () => $('#avatar-input').click());
    $('#avatar-input').addEventListener('change', handleAvatarChange);
  } catch (error) {
    showStatus(error.message);
  }
})();
