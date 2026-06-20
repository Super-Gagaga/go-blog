const API_BASE = '/api/v1';

const state = {
  page: 1,
  pageSize: 8,
  keyword: '',
  tagId: '',
  categoryId: '',
  sort: 'hot',
  articles: [],
  pagination: { total: 0, total_pages: 0 },
  authMode: 'login'
};

const els = {
  status: document.querySelector('#status-line'),
  featured: document.querySelector('#featured'),
  articleList: document.querySelector('#article-list'),
  feedTabs: document.querySelector('#feed-tabs'),
  categoryNav: document.querySelector('#category-nav'),
  topicList: document.querySelector('#topic-list'),
  authorList: document.querySelector('#author-list'),
  digestList: document.querySelector('#digest-list'),
  total: document.querySelector('#article-total'),
  title: document.querySelector('#section-title'),
  loadMore: document.querySelector('#load-more'),
  searchInput: document.querySelector('#search-input'),
  authArea: document.querySelector('#auth-area'),
  loginModal: document.querySelector('#login-modal'),
  loginPanel: document.querySelector('.login-panel'),
  loginForm: document.querySelector('#login-form'),
  loginEmail: document.querySelector('#login-email'),
  loginPassword: document.querySelector('#login-password'),
  registerUsername: document.querySelector('#register-username'),
  loginTitle: document.querySelector('#login-title'),
  loginSubtitle: document.querySelector('#login-subtitle'),
  loginError: document.querySelector('#login-error'),
  loginClose: document.querySelector('#login-close'),
  authModeToggle: document.querySelector('#auth-mode-toggle'),
  authSwitchText: document.querySelector('#auth-switch-text')
};

function showStatus(message) {
  els.status.textContent = message;
  els.status.classList.toggle('visible', Boolean(message));
}

async function getJSON(path, params = {}) {
  const url = new URL(API_BASE + path, window.location.origin);
  Object.entries(params).forEach(([key, value]) => {
    if (value !== '' && value !== null && value !== undefined) url.searchParams.set(key, value);
  });
  const res = await fetch(url);
  if (!res.ok) throw new Error(`请求失败：${res.status}`);
  const body = await res.json();
  if (body.code !== 0) throw new Error(body.message || '接口返回异常');
  return body.data;
}

function escapeHTML(value = '') {
  return String(value).replace(/[&<>"']/g, ch => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
  }[ch]));
}

function token() {
  return localStorage.getItem('jwt_token');
}

function avatarFor(user) {
  if (user && user.avatar) return user.avatar;
  const seed = encodeURIComponent((user && user.username) || 'writer');
  return `https://picsum.photos/seed/${seed}/100/100`;
}

function coverFor(article, wide = false) {
  if (article.cover_image) return article.cover_image;
  const seed = encodeURIComponent(article.slug || article.title || article.id || 'journal');
  return `https://picsum.photos/seed/${seed}/${wide ? '520/360' : '240/176'}`;
}

function formatDate(value) {
  if (!value) return '刚刚';
  return new Intl.DateTimeFormat('zh-CN', { month: 'short', day: 'numeric' }).format(new Date(value));
}

function readMinutes(article) {
  const text = article.summary || article.content || '';
  return Math.max(1, Math.ceil(text.length / 450));
}

function categoryName(article) {
  if (article.category && article.category.name) return article.category.name;
  if (article.tags && article.tags.length) return article.tags[0].name;
  return '随笔';
}

function openArticle(article) {
  window.location.href = `/article-detail?id=${article.id}`;
}

async function loadCurrentUser() {
  if (!token()) {
    renderLoggedOut();
    return;
  }
  try {
    const res = await fetch(`${API_BASE}/user/profile`, {
      headers: { Authorization: `Bearer ${token()}` }
    });
    if (res.status === 401) {
      localStorage.removeItem('jwt_token');
      localStorage.removeItem('refresh_token');
      renderLoggedOut();
      return;
    }
    const body = await res.json();
    if (body.code === 0 && body.data) renderLoggedIn(body.data);
    else renderLoggedOut();
  } catch {
    renderLoggedOut();
  }
}

function renderLoggedOut() {
  els.authArea.innerHTML = '<button class="btn-login" type="button" id="btn-login">登录</button>';
}

function renderLoggedIn(profile) {
  els.authArea.innerHTML = `
    <div class="user-dropdown" id="user-dropdown">
      <button class="user-menu" type="button" id="user-menu-button" aria-haspopup="menu" aria-expanded="false" title="${escapeHTML(profile.username || '我的主页')}">
        <img class="user-menu-avatar" src="${escapeHTML(avatarFor(profile))}" alt="">
        <span class="user-menu-name">${escapeHTML(profile.username || '用户')}</span>
        <i class="ph ph-caret-down user-menu-caret"></i>
      </button>
      <div class="user-menu-list" role="menu">
        <a href="/my-profile" role="menuitem"><i class="ph ph-user-circle"></i>我的主页</a>
        <button type="button" id="btn-logout" role="menuitem"><i class="ph ph-sign-out"></i>退出登录</button>
      </div>
    </div>
  `;
}

function closeUserMenu() {
  const dropdown = document.querySelector('#user-dropdown');
  const button = document.querySelector('#user-menu-button');
  if (!dropdown) return;
  dropdown.classList.remove('open');
  button?.setAttribute('aria-expanded', 'false');
}

function toggleUserMenu() {
  const dropdown = document.querySelector('#user-dropdown');
  const button = document.querySelector('#user-menu-button');
  if (!dropdown) return;
  const open = !dropdown.classList.contains('open');
  dropdown.classList.toggle('open', open);
  button?.setAttribute('aria-expanded', String(open));
}

function logout() {
  localStorage.removeItem('jwt_token');
  localStorage.removeItem('refresh_token');
  closeUserMenu();
  renderLoggedOut();
}

function setAuthMode(mode) {
  state.authMode = mode;
  const isRegister = mode === 'register';
  els.loginPanel.classList.toggle('register-mode', isRegister);
  els.loginTitle.textContent = isRegister ? '注册 Journal' : '登录 Journal';
  els.loginSubtitle.textContent = isRegister ? '创建账号后即可继续写作与阅读。' : '使用邮箱和密码继续写作与阅读。';
  els.loginForm.querySelector('.login-submit').textContent = isRegister ? '注册并登录' : '登录';
  els.authSwitchText.textContent = isRegister ? '已有账号？' : '还没有账号？';
  els.authModeToggle.textContent = isRegister ? '登录' : '注册';
  els.registerUsername.required = isRegister;
  els.loginPassword.autocomplete = isRegister ? 'new-password' : 'current-password';
  els.loginError.textContent = '';
}

function openLoginModal() {
  setAuthMode('login');
  els.loginPassword.value = '';
  els.loginError.textContent = '';
  els.loginModal.classList.add('visible');
  els.loginModal.setAttribute('aria-hidden', 'false');
  setTimeout(() => els.loginEmail.focus(), 0);
}

function closeLoginModal() {
  els.loginModal.classList.remove('visible');
  els.loginModal.setAttribute('aria-hidden', 'true');
  document.querySelector('#btn-login')?.focus();
}

async function login(email, password) {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || '登录失败，请检查邮箱和密码');
  localStorage.setItem('jwt_token', body.data.access_token);
  localStorage.setItem('refresh_token', body.data.refresh_token);
  closeLoginModal();
  await loadCurrentUser();
}

async function register(username, email, password) {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password })
  });
  const body = await res.json().catch(() => ({}));
  if (!res.ok || body.code !== 0) throw new Error(body.message || '注册失败，请检查填写信息');
  await login(email, password);
}

function articleMeta(article) {
  const user = article.user || {};
  return `
    <span>${escapeHTML(user.username || '匿名作者')}</span>
    <span class="dot"></span>
    <span>${formatDate(article.created_at)}</span>
    <span class="dot"></span>
    <span>${readMinutes(article)} 分钟阅读</span>
    <span class="dot"></span>
    <span class="article-stat"><i class="ph ph-chat-circle"></i>${article.comment_count || 0}</span>
    <span class="article-stat"><i class="ph ph-heart"></i>${article.like_count || 0}</span>
  `;
}

function renderArticles() {
  const [featuredArticle, ...rest] = state.articles;
  els.total.textContent = state.pagination.total || state.articles.length;
  if (!featuredArticle) {
    els.featured.innerHTML = '';
    els.articleList.innerHTML = '';
    showStatus('暂时没有已发布文章。');
    return;
  }

  showStatus('');
  els.featured.innerHTML = `
    <article class="featured-card" data-id="${featuredArticle.id}">
      <div>
        <span class="featured-tag">${escapeHTML(categoryName(featuredArticle))}</span>
        <h2 class="featured-title">${escapeHTML(featuredArticle.title)}</h2>
        <p class="featured-excerpt">${escapeHTML(featuredArticle.summary || '这篇文章还没有摘要。')}</p>
        <div class="featured-meta">
          <img src="${escapeHTML(avatarFor(featuredArticle.user))}" alt="" class="author-avatar">
          ${articleMeta(featuredArticle)}
        </div>
      </div>
      <img src="${escapeHTML(coverFor(featuredArticle, true))}" alt="" class="featured-img">
    </article>
  `;
  els.featured.querySelector('.featured-card').addEventListener('click', () => openArticle(featuredArticle));

  els.articleList.innerHTML = rest.map((article, index) => `
    <article class="article-item" data-id="${article.id}">
      <div class="article-body">
        <div class="article-rank">${String(index + 1).padStart(2, '0')}</div>
        <h3 class="article-title">${escapeHTML(article.title)}</h3>
        <p class="article-excerpt">${escapeHTML(article.summary || '这篇文章还没有摘要。')}</p>
        <div class="article-footer">
          <span class="inline-tag">${escapeHTML(categoryName(article))}</span>
          ${articleMeta(article)}
        </div>
      </div>
      <img src="${escapeHTML(coverFor(article))}" alt="" class="article-thumb">
    </article>
  `).join('');
  els.articleList.querySelectorAll('.article-item').forEach((item, index) => {
    item.addEventListener('click', () => openArticle(rest[index]));
  });
  renderDigest();
}

function renderCategories(categories) {
  els.feedTabs.innerHTML = '<button class="feed-tab active" type="button" data-category-id="">综合</button>' +
    categories.map(c => `<button class="feed-tab" type="button" data-category-id="${c.id}">${escapeHTML(c.name)}</button>`).join('');
  els.categoryNav.innerHTML = categories.map(c => `
    <li><a href="#" data-category-id="${c.id}"><i class="ph ph-folder-simple"></i>${escapeHTML(c.name)}</a></li>
  `).join('') || '<li><a href="#"><i class="ph ph-folder-simple"></i>暂无专题</a></li>';

  document.querySelectorAll('[data-category-id]').forEach(item => {
    item.addEventListener('click', event => {
      event.preventDefault();
      state.categoryId = item.dataset.categoryId || '';
      state.sort = '';
      state.tagId = '';
      state.keyword = '';
      state.page = 1;
      state.articles = [];
      setActiveNav('', state.categoryId);
      document.querySelectorAll('.feed-tab').forEach(tab => tab.classList.toggle('active', tab.dataset.categoryId === state.categoryId));
      els.title.textContent = state.categoryId ? item.textContent.trim() : '热门推荐';
      loadArticles();
    });
  });
}

async function renderTopics() {
  const topics = await getJSON('/topics/trending').catch(() => []);
  els.topicList.innerHTML = topics.slice(0, 12).map(t => `
    <a href="#" class="topic-pill" data-tag-id="${t.tag_id || ''}">${escapeHTML(t.name)}</a>
  `).join('') || '<span class="digest-meta">暂无话题数据</span>';
  els.topicList.querySelectorAll('[data-tag-id]').forEach(item => {
    item.addEventListener('click', event => {
      event.preventDefault();
      state.tagId = item.dataset.tagId;
      state.sort = '';
      state.categoryId = '';
      state.keyword = '';
      state.page = 1;
      state.articles = [];
      setActiveNav('', '');
      els.title.textContent = `话题：${item.textContent.trim()}`;
      loadArticles();
    });
  });
}

async function renderAuthors() {
  const authors = await getJSON('/recommendations/authors').catch(() => []);
  els.authorList.innerHTML = authors.slice(0, 5).map(author => `
    <div class="author-card">
      <img src="${escapeHTML(avatarFor(author))}" alt="" class="author-avatar-lg">
      <div class="author-info">
        <div class="author-name">${escapeHTML(author.username)}</div>
        <div class="author-bio">${escapeHTML(author.bio || `已发布 ${author.article_count || 0} 篇文章`)}</div>
      </div>
      <button class="btn-follow" type="button" data-user-id="${author.id}">关注</button>
    </div>
  `).join('') || '<div class="digest-meta">暂无推荐作者</div>';
}

async function renderDigest() {
  const picks = await getJSON('/articles/ranking', { period: 'week', limit: 5 }).catch(() => []);
  const list = picks.length ? picks : state.articles.slice(0, 4);
  els.digestList.innerHTML = list.map((article, index) => `
    <div class="digest-item" data-id="${article.id}">
      <span class="digest-num">${String(index + 1).padStart(2, '0')}</span>
      <div>
        <div class="digest-title">${escapeHTML(article.title)}</div>
        <div class="digest-meta">${article.view_count || 0} 阅读 · ${article.favorite_count || 0} 收藏</div>
      </div>
    </div>
  `).join('') || '<span class="digest-meta">暂无精选文章</span>';
  els.digestList.querySelectorAll('.digest-item').forEach((item, index) => item.addEventListener('click', () => openArticle(list[index])));
}

async function loadArticles(append = false) {
  showStatus('正在加载内容...');
  try {
    const result = await getJSON('/articles', {
      page: state.page,
      page_size: state.pageSize,
      status: 'published',
      category_id: state.categoryId,
      tag_id: state.tagId,
      keyword: state.keyword,
      sort: state.sort !== 'favorites' ? state.sort : 'hot'
    });
    state.pagination = result.pagination || state.pagination;
    state.articles = append ? state.articles.concat(result.list || []) : (result.list || []);
    renderArticles();
    els.loadMore.style.visibility = state.page < (state.pagination.total_pages || 0) ? 'visible' : 'hidden';
  } catch (error) {
    showStatus(error.message);
  }
}

async function loadFavorites(append = false) {
  if (!token()) {
    openLoginModal();
    return;
  }
  showStatus('正在加载收藏...');
  try {
    const res = await fetch(`${API_BASE}/user/favorites?page=${state.page}&page_size=${state.pageSize}`, {
      headers: { Authorization: `Bearer ${token()}` }
    });
    if (res.status === 401) {
      localStorage.removeItem('jwt_token');
      localStorage.removeItem('refresh_token');
      renderLoggedOut();
      showStatus('请先登录后再查看收藏');
      return;
    }
    const body = await res.json();
    if (body.code !== 0) throw new Error(body.message || '获取收藏失败');
    const data = body.data;
    state.pagination = data.pagination || state.pagination;
    const articles = (data.list || []).map(item => item.article).filter(Boolean);
    state.articles = append ? state.articles.concat(articles) : articles;
    renderArticles();
    els.loadMore.style.visibility = state.page < (state.pagination.total_pages || 0) ? 'visible' : 'hidden';
    if (!articles.length && !append) showStatus('还没有收藏任何文章');
  } catch (error) {
    showStatus(error.message);
  }
}

function setActiveNav(sort, categoryId) {
  document.querySelectorAll('.left-nav .nav-list li').forEach(li => li.classList.remove('active'));
  if (sort) {
    const link = document.querySelector(`.left-nav [data-sort="${sort}"]`);
    if (link) link.parentElement.classList.add('active');
  }
  if (categoryId) {
    const link = document.querySelector(`.left-nav [data-category-id="${categoryId}"]`);
    if (link) link.parentElement.classList.add('active');
  }
}

async function bootstrap() {
  const [categories] = await Promise.all([getJSON('/categories').catch(() => [])]);
  renderCategories(categories || []);
  renderTopics();
  renderAuthors();
  loadArticles();
}

/* Left-nav sort links */
document.querySelector('.left-nav').addEventListener('click', event => {
  const sortLink = event.target.closest('[data-sort]');
  if (!sortLink) return;
  event.preventDefault();
  const sort = sortLink.dataset.sort;
  state.sort = sort;
  state.categoryId = '';
  state.tagId = '';
  state.keyword = '';
  state.page = 1;
  state.articles = [];
  setActiveNav(sort, '');
  els.title.textContent = { hot: '热门推荐', latest: '最新文章', trend: '趋势文章', favorites: '我的收藏' }[sort] || '热门推荐';
  if (sort === 'favorites') {
    loadFavorites();
  } else {
    loadArticles();
  }
});

document.querySelector('#search-form').addEventListener('submit', event => {
  event.preventDefault();
  state.keyword = els.searchInput.value.trim();
  state.tagId = '';
  state.categoryId = '';
  state.sort = state.keyword ? '' : 'hot';
  state.page = 1;
  state.articles = [];
  setActiveNav(state.sort, '');
  els.title.textContent = state.keyword ? `搜索：${state.keyword}` : '热门推荐';
  loadArticles();
});

els.loadMore.addEventListener('click', () => {
  if (state.page < (state.pagination.total_pages || 0)) {
    state.page += 1;
    if (state.sort === 'favorites') {
      loadFavorites(true);
    } else {
      loadArticles(true);
    }
  }
});

document.addEventListener('click', async event => {
  if (event.target.closest('#btn-login')) {
    openLoginModal();
    return;
  }
  if (event.target.closest('#user-menu-button')) {
    event.stopPropagation();
    toggleUserMenu();
    return;
  }
  if (event.target.closest('#btn-logout')) {
    event.stopPropagation();
    logout();
    return;
  }
  if (!event.target.closest('#user-dropdown')) {
    closeUserMenu();
  }

  const navLink = event.target.closest('[data-nav]');
  if (navLink) {
    event.preventDefault();
    if (!token()) {
      openLoginModal();
      return;
    }
    const pages = { following: '/following', notifications: '/notifications', favorites: '/favorites', editor: '/editor' };
    window.location.href = pages[navLink.dataset.nav] || '/';
    return;
  }

  const follow = event.target.closest('.btn-follow');
  if (!follow) return;
  event.stopPropagation();
  if (!token()) {
    openLoginModal();
    return;
  }
  try {
    const res = await fetch(`${API_BASE}/users/${follow.dataset.userId}/follow`, {
      method: 'POST',
      headers: { Authorization: `Bearer ${token()}`, 'Content-Type': 'application/json' }
    });
    const body = await res.json();
    if (body.code !== 0) throw new Error(body.message || '操作失败');
    const following = body.data && body.data.following;
    follow.textContent = following ? '已关注' : '关注';
    follow.classList.toggle('active', Boolean(following));
  } catch (error) {
    alert(error.message);
  }
});

document.addEventListener('keydown', event => {
  if ((event.metaKey || event.ctrlKey) && event.key.toLowerCase() === 'k') {
    event.preventDefault();
    els.searchInput.focus();
  }
  if (event.key === 'Escape' && els.loginModal.classList.contains('visible')) closeLoginModal();
  if (event.key === 'Escape') closeUserMenu();
});

els.loginClose.addEventListener('click', closeLoginModal);
els.loginModal.addEventListener('click', event => {
  if (event.target === els.loginModal) closeLoginModal();
});
els.authModeToggle.addEventListener('click', () => setAuthMode(state.authMode === 'login' ? 'register' : 'login'));
els.loginForm.addEventListener('submit', async event => {
  event.preventDefault();
  const submit = els.loginForm.querySelector('.login-submit');
  submit.disabled = true;
  els.loginError.textContent = '';
  try {
    if (state.authMode === 'register') {
      await register(els.registerUsername.value.trim(), els.loginEmail.value.trim(), els.loginPassword.value);
    } else {
      await login(els.loginEmail.value.trim(), els.loginPassword.value);
    }
  } catch (error) {
    els.loginError.textContent = error.message;
  } finally {
    submit.disabled = false;
  }
});

loadCurrentUser();
bootstrap();
