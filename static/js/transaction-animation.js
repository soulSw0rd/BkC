// Animation fluide pour les transactions entre deux éléments
class TransactionAnimator {
    constructor() {
      this.pendingAnimations = [];
      this.isAnimating = false;
    }
    
    /**
     * Anime une transaction entre deux éléments
     * @param {string} fromElementId - ID de l'élément source
     * @param {string} toElementId - ID de l'élément destination
     * @param {number} amount - Montant de la transaction
     * @param {object} options - Options supplémentaires (couleur, durée...)
     */
    animateTransaction(fromElementId, toElementId, amount, options = {}) {
      // Paramètres par défaut
      const settings = {
        color: '#6366f1',
        duration: 1.5,
        particleCount: Math.min(Math.floor(amount), 30), // Limiter à 30 particules
        particleSize: 8,
        trailLength: 5,
        ...options
      };
      
      // Ajouter à la file d'attente
      this.pendingAnimations.push({ fromElementId, toElementId, amount, settings });
      
      // Lancer l'animation si aucune n'est en cours
      if (!this.isAnimating) {
        this.processNextAnimation();
      }
    }
    
    /**
     * Traiter la prochaine animation dans la file d'attente
     */
    processNextAnimation() {
      if (this.pendingAnimations.length === 0) {
        this.isAnimating = false;
        return;
      }
      
      this.isAnimating = true;
      const { fromElementId, toElementId, amount, settings } = this.pendingAnimations.shift();
      
      // Obtenir les positions des éléments
      const fromElement = document.getElementById(fromElementId);
      const toElement = document.getElementById(toElementId);
      
      if (!fromElement || !toElement) {
        console.error(`Éléments non trouvés: ${fromElementId} ou ${toElementId}`);
        this.processNextAnimation();
        return;
      }
      
      const fromRect = fromElement.getBoundingClientRect();
      const toRect = toElement.getBoundingClientRect();
      
      // Créer un conteneur pour l'animation
      const animationContainer = document.createElement('div');
      animationContainer.style.position = 'fixed';
      animationContainer.style.top = '0';
      animationContainer.style.left = '0';
      animationContainer.style.width = '100%';
      animationContainer.style.height = '100%';
      animationContainer.style.pointerEvents = 'none';
      animationContainer.style.zIndex = '9999';
      document.body.appendChild(animationContainer);
      
      // Créer les particules
      const particles = [];
      
      for (let i = 0; i < settings.particleCount; i++) {
        const particle = document.createElement('div');
        particle.style.position = 'absolute';
        particle.style.width = `${settings.particleSize}px`;
        particle.style.height = `${settings.particleSize}px`;
        particle.style.backgroundColor = settings.color;
        particle.style.borderRadius = '50%';
        particle.style.boxShadow = `0 0 ${settings.trailLength}px ${settings.color}`;
        particle.style.opacity = '0.8';
        
        // Position initiale (légèrement aléatoire autour du point de départ)
        const offset = 20;
        const startX = fromRect.left + fromRect.width / 2 + (Math.random() * offset * 2 - offset);
        const startY = fromRect.top + fromRect.height / 2 + (Math.random() * offset * 2 - offset);
        
        particle.style.left = `${startX}px`;
        particle.style.top = `${startY}px`;
        
        animationContainer.appendChild(particle);
        particles.push(particle);
        
        // Ajouter un délai pour chaque particule
        const delay = (i / settings.particleCount) * 0.5; // Répartir sur 0.5s
        
        // Animation avec courbe de Bézier pour un mouvement naturel
        setTimeout(() => {
          // Points de contrôle pour la courbe de Bézier
          const controlPoint1X = startX + (Math.random() * 100 - 50);
          const controlPoint1Y = startY - 100 - (Math.random() * 100);
          const controlPoint2X = toRect.left + toRect.width / 2 + (Math.random() * 100 - 50);
          const controlPoint2Y = toRect.top + toRect.height / 2 - 100 - (Math.random() * 100);
          const endX = toRect.left + toRect.width / 2;
          const endY = toRect.top + toRect.height / 2;
          
          // Appliquer l'animation
          particle.style.transition = `left ${settings.duration}s cubic-bezier(.25,.75,.5,1.25), top ${settings.duration}s cubic-bezier(.25,.75,.5,1.25), opacity 0.3s ease-in-out`;
          
          // Animation keyframes avec requestAnimationFrame pour plus de fluidité
          const startTime = performance.now();
          const duration = settings.duration * 1000; // Convertir en millisecondes
          
          function animateParticle(timestamp) {
            const elapsedTime = timestamp - startTime;
            const progress = Math.min(elapsedTime / duration, 1);
            
            if (progress < 1) {
              // Calcul de la position sur la courbe de Bézier cubique
              const t = progress;
              const u = 1 - t;
              const tt = t * t;
              const uu = u * u;
              const uuu = uu * u;
              const ttt = tt * t;
              
              const x = uuu * startX + 3 * uu * t * controlPoint1X + 3 * u * tt * controlPoint2X + ttt * endX;
              const y = uuu * startY + 3 * uu * t * controlPoint1Y + 3 * u * tt * controlPoint2Y + ttt * endY;
              
              particle.style.left = `${x}px`;
              particle.style.top = `${y}px`;
              
              requestAnimationFrame(animateParticle);
            } else {
              // Animation terminée
              particle.style.left = `${endX}px`;
              particle.style.top = `${endY}px`;
              particle.style.opacity = '0';
              
              // Vérifier si c'est la dernière particule
              if (i === settings.particleCount - 1) {
                // Nettoyer après l'animation
                setTimeout(() => {
                  animationContainer.remove();
                  this.processNextAnimation();
                }, 300); // Petit délai pour l'animation de fadeout
              }
            }
          }
          
          requestAnimationFrame(animateParticle);
        }, delay * 1000);
      }
      
      // Créer une notification flottante avec le montant
      const amountDisplay = document.createElement('div');
      amountDisplay.textContent = `${amount.toFixed(2)} BCK`;
      amountDisplay.style.position = 'absolute';
      amountDisplay.style.left = `${fromRect.left + fromRect.width / 2}px`;
      amountDisplay.style.top = `${fromRect.top - 20}px`;
      amountDisplay.style.transform = 'translate(-50%, -100%)';
      amountDisplay.style.backgroundColor = 'rgba(99, 102, 241, 0.9)';
      amountDisplay.style.color = 'white';
      amountDisplay.style.padding = '5px 10px';
      amountDisplay.style.borderRadius = '10px';
      amountDisplay.style.fontWeight = 'bold';
      amountDisplay.style.opacity = '0';
      amountDisplay.style.transition = 'opacity 0.3s ease-in-out, transform 0.3s ease-out';
      
      animationContainer.appendChild(amountDisplay);
      
      // Afficher l'animation du montant
      setTimeout(() => {
        amountDisplay.style.opacity = '1';
        amountDisplay.style.transform = 'translate(-50%, -120%)';
      }, 100);
      
      // Faire disparaître l'animation du montant
      setTimeout(() => {
        amountDisplay.style.opacity = '0';
        amountDisplay.style.transform = 'translate(-50%, -150%)';
      }, 1000);
    }
  }
  
  // Créer une instance globale
  window.transactionAnimator = new TransactionAnimator();
  
  // Exemple d'utilisation:
  // window.transactionAnimator.animateTransaction('wallet-icon', 'recipient-wallet', 25.5, { color: '#6366f1' });