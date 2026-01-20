# ModulOps Frontend

Application React + TypeScript pour la plateforme DevOps ModulOps.

## Technologies

- **React 18** : Framework UI
- **TypeScript 5** : Typage statique
- **Vite** : Build tool et serveur de développement
- **Material-UI (MUI)** : Bibliothèque de composants UI
- **React Query** : Gestion des données et cache
- **React Router** : Routage
- **Socket.io Client** : WebSocket pour les événements temps réel
- **Recharts** : Visualisations de données
- **Zustand** : Gestion d'état légère
- **Axios** : Client HTTP

## Installation

```bash
cd frontend
npm install
```

## Développement

```bash
npm run dev
```

L'application sera accessible sur http://localhost:5173

## Build

```bash
npm run build
```

Les fichiers de production seront générés dans le dossier `dist/`.

## Structure

```
frontend/
├── src/
│   ├── components/      # Composants réutilisables
│   │   └── Layout.tsx    # Layout principal avec navigation
│   ├── contexts/         # Contextes React
│   │   ├── AuthContext.tsx    # Gestion de l'authentification
│   │   └── SocketContext.tsx  # Gestion WebSocket
│   ├── pages/           # Pages de l'application
│   │   ├── LoginPage.tsx
│   │   ├── RegisterPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── K8sPage.tsx
│   │   ├── TerraformPage.tsx
│   │   ├── AnsiblePage.tsx
│   │   ├── PipelinePage.tsx
│   │   ├── AlertsPage.tsx
│   │   ├── MetricsPage.tsx
│   │   └── SettingsPage.tsx
│   ├── services/        # Services API
│   │   ├── api.ts       # Client API de base
│   │   ├── authService.ts
│   │   └── k8sService.ts
│   ├── App.tsx          # Composant racine avec routes
│   ├── main.tsx         # Point d'entrée
│   ├── theme.ts         # Configuration Material-UI
│   └── index.css        # Styles globaux
├── package.json
├── vite.config.ts
└── tsconfig.json
```

## Fonctionnalités

### Authentification
- Connexion / Inscription
- Gestion des tokens JWT
- Rafraîchissement automatique des tokens
- Protection des routes privées

### Dashboard
- Vue d'ensemble des services
- Statistiques en temps réel
- Événements récents via WebSocket

### Kubernetes
- Liste des namespaces
- Visualisation des pods par namespace
- Statuts en temps réel

### WebSocket
- Connexion automatique après authentification
- Réception d'événements K8s, Terraform, Pipelines, Alertes
- Indicateur de statut de connexion

### Visualisations
- Graphiques avec Recharts
- Métriques CPU/Mémoire
- Requêtes par jour

## Configuration

Les variables d'environnement peuvent être définies dans un fichier `.env` :

```env
VITE_API_BASE_URL=http://localhost:8000
VITE_SOCKET_URL=http://localhost:8000
```

## Notes

- Le frontend communique avec l'API Gateway Kong sur le port 8000
- L'authentification utilise JWT stocké dans localStorage
- Les événements temps réel sont reçus via WebSocket (Socket.io)
- Les pages Terraform, Ansible et Pipelines sont des placeholders en attendant l'implémentation des services backend
