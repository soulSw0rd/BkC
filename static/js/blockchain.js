// Animation et fonctionnalités pour la page blockchain
document.addEventListener('DOMContentLoaded', function() {
  // Animation d'entrée pour les cartes de blocs
  const blockCards = document.querySelectorAll('.block-card');
  blockCards.forEach((card, index) => {
    card.style.opacity = '0';
    card.style.transform = 'translateY(20px)';
    card.style.transition = 'opacity 0.3s ease, transform 0.3s ease';
    
    setTimeout(() => {
      card.style.opacity = '1';
      card.style.transform = 'translateY(0)';
    }, 100 + (index * 100)); // Délai progressif pour chaque carte
  });
  
  // Affichage des informations de minage
  displayMiningInfo();
  
  // Visualisation graphique de la chaîne de blocs
  createBlockchainVisualization();
  
  // Affichage des hashs récents
  displayRecentHashes();
  
  // Gestion du formulaire d'ajout de hash
  setupHashForm();
  
  // Charger et afficher le classement des mineurs
  loadMinersLeaderboard();
  
  // Fonction pour afficher les hashs récents
  function displayRecentHashes() {
    const recentHashesContainer = document.getElementById('recent-hashes');
    if (!recentHashesContainer) return;
    
    // Récupérer les 5 derniers blocs (sauf le genesis)
    const recentBlocks = Array.from(document.querySelectorAll('.block-card')).slice(1, 6);
    
    if (recentBlocks.length === 0) {
      recentHashesContainer.innerHTML = '<p class="text-gray-400">Aucun hash disponible pour le moment.</p>';
      return;
    }
    
    // Vider le conteneur
    recentHashesContainer.innerHTML = '';
    
    // Créer un élément pour chaque hash récent
    recentBlocks.forEach(blockCard => {
      const blockIndex = blockCard.querySelector('h3').textContent.replace('Bloc #', '');
      const blockHash = blockCard.querySelector('.font-mono').textContent;
      const blockMiner = blockCard.querySelector('.text-amber-400')?.textContent.replace('⛏️ Miné par: ', '') || 'Anonyme';
      
      const hashElement = document.createElement('div');
      hashElement.className = 'bg-gray-700 rounded p-3 flex justify-between items-center';
      hashElement.innerHTML = `
        <div class="flex-1">
          <div class="flex items-center">
            <span class="font-bold text-amber-400 mr-2">#${blockIndex}</span>
            <span class="text-xs text-gray-400">par</span>
            <span class="font-semibold text-blue-400 ml-2">${blockMiner}</span>
          </div>
          <div class="font-mono text-xs text-gray-300 truncate">${blockHash}</div>
        </div>
        <div class="flex-shrink-0">
          <div class="text-xs text-gray-400">
            <span class="bg-blue-900/50 px-2 py-1 rounded">
              Hash ajouté 
            </span>
          </div>
        </div>
      `;
      
      recentHashesContainer.appendChild(hashElement);
    });
  }
  
  // Fonction pour configurer le formulaire d'ajout de hash
  function setupHashForm() {
    const form = document.getElementById('add-block-form');
    const miningStatus = document.getElementById('mining-status');
    const successMessage = document.getElementById('success-message');
    
    if (!form) return;
    
    form.addEventListener('submit', function(e) {
      e.preventDefault();
      
      const blockData = document.getElementById('block-data').value;
      if (!blockData.trim()) {
        alert('Veuillez entrer un message pour accompagner votre hash');
        return;
      }
      
      // Afficher le statut de génération
      miningStatus.classList.remove('hidden');
      successMessage.classList.add('hidden');
      
      // Envoyer la requête pour ajouter un bloc
      fetch('/blockchain', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ data: blockData })
      })
      .then(response => response.json())
      .then(data => {
        // Le serveur a répondu avec succès
        miningStatus.classList.add('hidden');
        successMessage.classList.remove('hidden');
        
        // Afficher un message avec des informations sur le hash
        successMessage.innerHTML = `
          <div class="flex items-center">
            <svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 20 20">
              <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
            </svg>
            <span>Votre hash a été ajouté à la blockchain avec succès! </span>
            <span class="ml-2 text-xs bg-green-900 rounded-full px-2 py-1">Miné par: ${data.miner || 'vous'}</span>
          </div>
        `;
        
        // Recharger la page après un court délai pour voir le nouveau bloc
        setTimeout(() => {
          window.location.reload();
        }, 2000);
      })
      .catch(error => {
        console.error('Erreur:', error);
        miningStatus.classList.add('hidden');
        alert('Erreur lors de l\'ajout du hash: ' + error.message);
      });
    });
  }
  
  // Fonction pour afficher les informations de minage
  function displayMiningInfo() {
    document.querySelectorAll('.mining-info').forEach(infoElement => {
      const miningInfoRaw = infoElement.getAttribute('data-mining-info');
      
      try {
        const miningInfo = JSON.parse(miningInfoRaw);
        
        // Créer l'affichage des informations de minage
        let infoHtml = `
          <div class="flex flex-wrap gap-3">
            <span class="inline-flex items-center bg-amber-900/50 px-2 py-1 rounded">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20"><path d="M10 9a3 3 0 100-6 3 3 0 000 6z"/><path fill-rule="evenodd" d="M10 0C4.477 0 0 4.484 0 10.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0110 4.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.203 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.942.359.31.678.921.678 1.856 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0020 10.017C20 4.484 15.522 0 10 0z" clip-rule="evenodd"/></svg>
              ${miningInfo.miner}
            </span>
            
            <span class="inline-flex items-center bg-gray-800 px-2 py-1 rounded">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd"/></svg>
              ${Math.round(miningInfo.duration_ms / 1000)} secondes
            </span>
            
            <span class="inline-flex items-center bg-gray-800 px-2 py-1 rounded">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd"/></svg>
              Difficulté: ${miningInfo.difficulty}
            </span>
            
            <span class="inline-flex items-center bg-gray-800 px-2 py-1 rounded">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2h2a1 1 0 100-2H9z" clip-rule="evenodd"/></svg>
              Nonce: ${miningInfo.nonce}
            </span>

            <span class="inline-flex items-center bg-gray-800 px-2 py-1 rounded">
              <svg class="w-4 h-4 mr-1" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M6 2a1 1 0 00-1 1v1H4a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V6a2 2 0 00-2-2h-1V3a1 1 0 10-2 0v1H7V3a1 1 0 00-1-1zm0 5a1 1 0 000 2h8a1 1 0 100-2H6z" clip-rule="evenodd"/></svg>
              ${new Date(miningInfo.timestamp).toLocaleString()}
            </span>
          </div>
        `;
        
        infoElement.innerHTML = infoHtml;
      } catch (e) {
        infoElement.innerHTML = `<div class="text-red-400">Erreur de décodage des informations de minage: ${e.message}</div>`;
      }
    });
  }
  
  // Fonction pour créer une visualisation de la blockchain
  function createBlockchainVisualization() {
    const blocks = Array.from(document.querySelectorAll('.block-card'));
    if (blocks.length <= 1) return; // Pas assez de blocs pour visualiser
    
    // Créer le conteneur de visualisation
    const visContainer = document.createElement('div');
    visContainer.className = 'my-12 p-4 bg-gray-800 rounded-lg shadow-lg overflow-x-auto';
    visContainer.innerHTML = '<h2 class="text-2xl font-bold text-blue-400 mb-6">Visualisation de la Blockchain</h2>';
    
    // Créer le conteneur pour le graphique
    const graphContainer = document.createElement('div');
    graphContainer.className = 'flex items-center justify-start min-w-max py-8';
    visContainer.appendChild(graphContainer);
    
    // Ajouter chaque bloc à la visualisation
    blocks.forEach((block, index) => {
      const blockIndex = block.querySelector('h3').textContent.replace('Bloc #', '');
      const blockHash = block.querySelector('.font-mono').textContent.substring(0, 8) + '...';
      
      // Créer l'élément pour ce bloc
      const blockElem = document.createElement('div');
      blockElem.className = 'flex flex-col items-center mx-4 first:ml-0';
      blockElem.innerHTML = `
        <div class="block-vis w-24 h-24 bg-blue-900 rounded-lg flex flex-col items-center justify-center text-center p-2 shadow-md">
          <div class="text-xs font-bold">#${blockIndex}</div>
          <div class="text-xs mt-1 font-mono">${blockHash}</div>
        </div>
      `;
      
      // Ajouter une flèche entre les blocs (sauf pour le premier)
      if (index > 0) {
        const arrow = document.createElement('div');
        arrow.className = 'arrow mr-4 text-blue-400 text-2xl';
        arrow.textContent = '←';
        graphContainer.appendChild(arrow);
      }
      
      graphContainer.appendChild(blockElem);
    });
    
    // Insérer la visualisation après l'en-tête de la page
    const header = document.querySelector('h1');
    header.parentNode.insertBefore(visContainer, header.nextSibling);
    
    // Ajouter une animation aux blocs dans la visualisation
    const blockVis = document.querySelectorAll('.block-vis');
    blockVis.forEach(block => {
      block.addEventListener('mouseenter', function() {
        this.classList.add('scale-110', 'shadow-lg');
        this.style.transition = 'transform 0.2s ease, box-shadow 0.2s ease';
        this.style.backgroundColor = 'rgba(30, 64, 175, 0.9)';
      });
      
      block.addEventListener('mouseleave', function() {
        this.classList.remove('scale-110', 'shadow-lg');
        this.style.backgroundColor = '';
      });
    });
  }
  
  // Animation pour le survol des cartes de bloc
  blockCards.forEach(card => {
    card.addEventListener('mouseenter', function() {
      this.classList.add('shadow-xl');
    });
    
    card.addEventListener('mouseleave', function() {
      this.classList.remove('shadow-xl');
    });
  });
  
  // Copie du hash dans le presse-papier lors du clic
  document.querySelectorAll('.font-mono').forEach(hashElem => {
    hashElem.title = 'Cliquez pour copier';
    hashElem.classList.add('cursor-pointer');
    
    hashElem.addEventListener('click', function() {
      const hash = this.textContent;
      navigator.clipboard.writeText(hash).then(() => {
        // Afficher une notification temporaire
        const notification = document.createElement('div');
        notification.className = 'fixed bottom-4 right-4 bg-green-500 text-white py-2 px-4 rounded shadow-lg transition-opacity duration-300';
        notification.textContent = 'Hash copié dans le presse-papier!';
        document.body.appendChild(notification);
        
        setTimeout(() => {
          notification.style.opacity = '0';
          setTimeout(() => notification.remove(), 300);
        }, 2000);
      });
    });
  });
  
  // Fonction pour charger et afficher le classement des mineurs
  function loadMinersLeaderboard() {
    const leaderboardContainer = document.getElementById('miners-leaderboard');
    if (!leaderboardContainer) return;
    
    // Récupérer les statistiques des mineurs
    fetch('/miners-stats')
      .then(response => {
        if (!response.ok) {
          throw new Error('Erreur lors de la récupération des statistiques des mineurs');
        }
        return response.json();
      })
      .then(miners => {
        // Vider le conteneur
        leaderboardContainer.innerHTML = '';
        
        if (miners.length === 0) {
          const emptyRow = document.createElement('tr');
          emptyRow.innerHTML = `
            <td colspan="4" class="py-4 px-4 text-center text-gray-400">
              Aucun mineur actif pour le moment.
            </td>
          `;
          leaderboardContainer.appendChild(emptyRow);
          return;
        }
        
        // Afficher chaque mineur dans le tableau
        miners.forEach((miner, index) => {
          const row = document.createElement('tr');
          
          // Formater la date du dernier minage
          let lastMiningDate = 'Jamais';
          if (miner.lastMiningTime > 0) {
            lastMiningDate = new Date(miner.lastMiningTime * 1000).toLocaleString();
          }
          
          // Déterminer la position (médaille)
          let positionDisplay = `${index + 1}`;
          if (index === 0) {
            positionDisplay = '🥇';
          } else if (index === 1) {
            positionDisplay = '🥈';
          } else if (index === 2) {
            positionDisplay = '🥉';
          }
          
          row.innerHTML = `
            <td class="py-2 px-4 border-b border-gray-600">
              <span class="font-bold">${positionDisplay}</span>
            </td>
            <td class="py-2 px-4 border-b border-gray-600">
              <span class="font-semibold text-blue-400">${miner.username}</span>
            </td>
            <td class="py-2 px-4 border-b border-gray-600">
              <span class="bg-amber-900/50 px-2 py-1 rounded">${miner.blocksMined}</span>
            </td>
            <td class="py-2 px-4 border-b border-gray-600 text-gray-300 text-sm">
              ${lastMiningDate}
            </td>
          `;
          
          leaderboardContainer.appendChild(row);
        });
      })
      .catch(error => {
        console.error('Erreur:', error);
        leaderboardContainer.innerHTML = `
          <tr>
            <td colspan="4" class="py-4 px-4 text-center text-red-400">
              Erreur lors du chargement des statistiques: ${error.message}
            </td>
          </tr>
        `;
      });
  }
}); 