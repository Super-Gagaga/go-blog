(() => {
  const STORAGE_KEY = 'journal_theme';
  const DARK = 'dark';
  const LIGHT = 'light';

  function storedTheme() {
    return localStorage.getItem(STORAGE_KEY) === DARK ? DARK : LIGHT;
  }

  function updateToggle(theme) {
    document.querySelectorAll('[data-theme-toggle]').forEach(button => {
      const isDark = theme === DARK;
      button.setAttribute('aria-pressed', String(isDark));
      const icon = button.querySelector('i');
      const label = button.querySelector('[data-theme-label]');
      if (icon) icon.className = `ph ${isDark ? 'ph-sun' : 'ph-moon'}`;
      if (label) label.textContent = isDark ? '白天模式' : '夜间模式';
      button.title = isDark ? '切换到白天模式' : '切换到夜间模式';
    });
  }

  function applyTheme(theme) {
    const nextTheme = theme === DARK ? DARK : LIGHT;
    document.documentElement.dataset.theme = nextTheme;
    document.documentElement.style.colorScheme = nextTheme === DARK ? DARK : LIGHT;
    updateToggle(nextTheme);
  }

  function setTheme(theme) {
    const nextTheme = theme === DARK ? DARK : LIGHT;
    localStorage.setItem(STORAGE_KEY, nextTheme);
    applyTheme(nextTheme);
  }

  applyTheme(storedTheme());

  document.addEventListener('DOMContentLoaded', () => {
    applyTheme(storedTheme());
    document.querySelectorAll('[data-theme-toggle]').forEach(button => {
      button.addEventListener('click', () => {
        setTheme(storedTheme() === DARK ? LIGHT : DARK);
      });
    });
  });

  window.addEventListener('storage', event => {
    if (event.key === STORAGE_KEY) applyTheme(storedTheme());
  });

  window.JournalTheme = { applyTheme, setTheme, storedTheme };
})();
