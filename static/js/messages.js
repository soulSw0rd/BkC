// Animation des éléments de la page messages
document.addEventListener('DOMContentLoaded', function() {
  // Animer l'affichage des messages au chargement de la page
  const messages = document.querySelectorAll('.message');
  messages.forEach((message, index) => {
    message.style.opacity = '0';
    message.style.transform = 'translateY(20px)';
    message.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    
    setTimeout(() => {
      message.style.opacity = '1';
      message.style.transform = 'translateY(0)';
    }, 100 + (index * 50)); // Délai progressif pour chaque message
  });

  // Animation du champ de recherche
  const searchInput = document.getElementById('search-user');
  if (searchInput) {
    searchInput.addEventListener('focus', function() {
      this.classList.add('ring-2', 'ring-blue-500');
    });
    
    searchInput.addEventListener('blur', function() {
      this.classList.remove('ring-2', 'ring-blue-500');
    });
    
    // Filtrer les conversations lors de la recherche
    searchInput.addEventListener('input', function() {
      const searchTerm = this.value.toLowerCase();
      const conversations = document.querySelectorAll('.conversation-item');
      
      conversations.forEach(conversation => {
        const username = conversation.querySelector('.font-bold').textContent.toLowerCase();
        if (username.includes(searchTerm)) {
          conversation.style.display = 'block';
        } else {
          conversation.style.display = 'none';
        }
      });
    });
  }

  // Animation de pulsation pour les indicateurs d'activité blockchain
  const blockchainIndicator = document.querySelector('.animate-spin');
  if (blockchainIndicator) {
    setInterval(() => {
      blockchainIndicator.classList.add('scale-110');
      setTimeout(() => {
        blockchainIndicator.classList.remove('scale-110');
      }, 500);
    }, 2000);
  }

  // Éviter la soumission du formulaire si le champ est vide
  const messageForm = document.getElementById('message-form');
  const messageContent = document.getElementById('message-content');
  
  if (messageForm && messageContent) {
    messageForm.addEventListener('submit', function(e) {
      if (!messageContent.value.trim()) {
        e.preventDefault();
        messageContent.classList.add('border-red-500');
        messageContent.classList.add('shake-animation');
        
        setTimeout(() => {
          messageContent.classList.remove('shake-animation');
        }, 500);
      }
    });
    
    // Retirer la bordure rouge quand l'utilisateur commence à taper
    messageContent.addEventListener('input', function() {
      this.classList.remove('border-red-500');
    });
  }
});

// Définition de l'animation de secousse
const style = document.createElement('style');
style.textContent = `
  @keyframes shake {
    0%, 100% { transform: translateX(0); }
    10%, 30%, 50%, 70%, 90% { transform: translateX(-5px); }
    20%, 40%, 60%, 80% { transform: translateX(5px); }
  }
  
  .shake-animation {
    animation: shake 0.5s cubic-bezier(.36,.07,.19,.97) both;
  }
  
  .message {
    transition: background-color 0.3s ease;
  }
  
  .sent:hover {
    background-color: rgba(30, 64, 175, 0.9);
  }
  
  .received:hover {
    background-color: rgba(55, 65, 81, 0.9);
  }
  
  #notification-area > div {
    transition: opacity 0.5s ease;
  }
`;
document.head.appendChild(style); 