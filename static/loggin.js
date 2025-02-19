document.addEventListener('DOMContentLoaded', () => {
    console.log('🔥 CryptoChain Go Login Loaded');

    // Effet de focus sur les champs
    document.querySelectorAll('input').forEach(input => {
        input.addEventListener('focus', () => {
            input.classList.add('ring-2', 'ring-purple-500');
        });
        input.addEventListener('blur', () => {
            input.classList.remove('ring-2', 'ring-purple-500');
        });
    });

    // Mode sombre / clair
    const themeToggle = document.getElementById('themeToggle');
    themeToggle.addEventListener('click', () => {
        document.body.classList.toggle('bg-white');
        document.body.classList.toggle('text-black');
    });

    // Animation du bouton de connexion
    const loginButton = document.querySelector('button[type="submit"]');
    loginButton.addEventListener('mouseover', () => {
        loginButton.classList.add('animate-pulse');
    });
    loginButton.addEventListener('mouseleave', () => {
        loginButton.classList.remove('animate-pulse');
    });
});
