// Gestionnaire de notifications
class NotificationManager {
    constructor() {
        this.container = document.createElement('div');
        this.container.id = 'notification-container';
        this.container.style.position = 'fixed';
        this.container.style.bottom = '20px';
        this.container.style.right = '20px';
        this.container.style.zIndex = '1000';
        document.body.appendChild(this.container);
    }

    show(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `notification notification-${type} fade-in`;
        notification.textContent = message;

        this.container.appendChild(notification);

        setTimeout(() => {
            notification.style.animation = 'slideOut 0.3s forwards';
            setTimeout(() => {
                this.container.removeChild(notification);
            }, 300);
        }, 3000);
    }
}

// Gestionnaire d'animations
const animate = {
    fadeIn: (element) => {
        element.style.opacity = '0';
        element.classList.add('fade-in');
        element.style.opacity = '1';
    },

    slideIn: (element) => {
        element.style.transform = 'translateX(100%)';
        element.style.transition = 'transform 0.3s ease-out';
        setTimeout(() => {
            element.style.transform = 'translateX(0)';
        }, 10);
    }
};

// Gestionnaire de formulaires
class FormManager {
    static async handleSubmit(form, endpoint) {
        try {
            const formData = new FormData(form);
            const data = Object.fromEntries(formData.entries());

            const response = await fetch(endpoint, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(data)
            });

            if (!response.ok) {
                throw new Error('Erreur lors de la soumission du formulaire');
            }

            return await response.json();
        } catch (error) {
            console.error('Erreur:', error);
            throw error;
        }
    }
}

// Gestionnaire de transactions
class TransactionManager {
    static async getTransactions() {
        try {
            const response = await fetch('/api/transactions');
            if (!response.ok) {
                throw new Error('Erreur lors de la récupération des transactions');
            }
            return await response.json();
        } catch (error) {
            console.error('Erreur:', error);
            throw error;
        }
    }

    static async createTransaction(transactionData) {
        try {
            const response = await fetch('/api/transactions', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(transactionData)
            });

            if (!response.ok) {
                throw new Error('Erreur lors de la création de la transaction');
            }

            return await response.json();
        } catch (error) {
            console.error('Erreur:', error);
            throw error;
        }
    }
}

// Initialisation
document.addEventListener('DOMContentLoaded', () => {
    // Création du gestionnaire de notifications
    window.notifications = new NotificationManager();

    // Animation des cartes au chargement
    document.querySelectorAll('.card').forEach(card => {
        animate.fadeIn(card);
    });

    // Gestion des formulaires
    document.querySelectorAll('form').forEach(form => {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            try {
                const endpoint = form.getAttribute('data-endpoint');
                const result = await FormManager.handleSubmit(form, endpoint);
                window.notifications.show('Opération réussie !', 'success');
            } catch (error) {
                window.notifications.show('Une erreur est survenue', 'error');
            }
        });
    });
}); 