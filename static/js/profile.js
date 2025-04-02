document.addEventListener('DOMContentLoaded', () => {
    console.log('🧩 CryptoChain Go - Profil Utilisateur Chargé');
    
    // Animation du titre
    const title = document.querySelector('h1');
    if (title) {
      title.classList.add('animate-pulse');
      setTimeout(() => {
        title.classList.remove('animate-pulse');
      }, 2000);
    }
    
    // Hover effect sur les boutons
    const buttons = document.querySelectorAll('button, a.bg-purple-500, a.bg-blue-500');
    buttons.forEach(button => {
      button.addEventListener('mouseover', () => {
        button.classList.add('shadow-lg', 'shadow-purple-500/20');
      });
      button.addEventListener('mouseleave', () => {
        button.classList.remove('shadow-lg', 'shadow-purple-500/20');
      });
    });
  });