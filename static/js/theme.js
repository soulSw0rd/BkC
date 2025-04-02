document.addEventListener('DOMContentLoaded', () => {
  // Vérifier la préférence utilisateur stockée ou utiliser les préférences système
  const isDarkMode = localStorage.getItem('darkMode') === 'true' || 
                    (window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches);
  
  // Appliquer le thème initial
  document.documentElement.classList.toggle('dark', isDarkMode);
  updateThemeIcon(isDarkMode);
  
  // Configurer le bouton de basculement
  const themeToggle = document.getElementById('themeToggle');
  if (themeToggle) {
    themeToggle.addEventListener('click', () => {
      const isDark = document.documentElement.classList.toggle('dark');
      localStorage.setItem('darkMode', isDark);
      updateThemeIcon(isDark);
      
      // Animation de transition
      document.documentElement.style.transition = 'background-color 0.5s ease, color 0.5s ease';
      setTimeout(() => {
        document.documentElement.style.transition = '';
      }, 500);
    });
  }
});

function updateThemeIcon(isDark) {
  const themeToggle = document.getElementById('themeToggle');
  if (themeToggle) {
    themeToggle.innerHTML = isDark ? '☀️' : '🌙';
    themeToggle.setAttribute('title', isDark ? 'Passer au mode clair' : 'Passer au mode sombre');
  }
}