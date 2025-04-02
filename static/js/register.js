document.addEventListener('DOMContentLoaded', () => {
    console.log('🔐 CryptoChain Go - Registration Loaded');
    
    // Évaluation de la force du mot de passe
    const passwordInput = document.getElementById('password');
    let passwordStrength = 0;
    
    passwordInput.addEventListener('input', () => {
      const password = passwordInput.value;
      let strength = 0;
      
      // Longueur
      if (password.length >= 8) strength += 1;
      if (password.length >= 12) strength += 1;
      
      // Caractères spéciaux
      if (/[!@#$%^&*(),.?":{}|<>]/.test(password)) strength += 1;
      
      // Chiffres
      if (/\d/.test(password)) strength += 1;
      
      // Majuscules et minuscules
      if (/[a-z]/.test(password) && /[A-Z]/.test(password)) strength += 1;
      
      passwordStrength = strength;
      
      // Afficher visuellement la force
      let strengthClass = 'bg-red-500';
      if (strength >= 4) strengthClass = 'bg-green-500';
      else if (strength >= 2) strengthClass = 'bg-yellow-500';
      
      // Ajouter un indicateur visuel si pas déjà présent
      let strengthIndicator = document.getElementById('password-strength');
      if (!strengthIndicator) {
        strengthIndicator = document.createElement('div');
        strengthIndicator.id = 'password-strength';
        strengthIndicator.classList.add('h-1', 'mt-1', 'rounded', 'transition-all', 'duration-300');
        passwordInput.parentElement.appendChild(strengthIndicator);
      }
      
      // Mise à jour de l'indicateur
      strengthIndicator.className = `h-1 mt-1 rounded transition-all duration-300 ${strengthClass}`;
      strengthIndicator.style.width = `${(strength / 5) * 100}%`;
    });
  });