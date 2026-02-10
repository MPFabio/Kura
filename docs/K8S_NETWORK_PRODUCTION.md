# Problèmes Réseau K8s en Production - Solutions

## Problèmes identifiés

### 1. Endpoints locaux en production
- Les kubeconfigs peuvent contenir des endpoints locaux (`127.0.0.1`, `localhost`, `host.docker.internal`)
- Ces endpoints ne fonctionnent pas en production où les clusters sont distants

### 2. Sécurité TLS
- Le code forçait `InsecureSkipTLSVerify = true` pour les clusters locaux
- En production, cela pose un risque de sécurité majeur

### 3. Timeouts et retry
- Pas de gestion de timeouts configurables pour les connexions distantes
- Pas de mécanisme de retry pour les erreurs réseau temporaires

### 4. Validation des endpoints
- Aucune validation pour rejeter les endpoints locaux en production

## Solutions implémentées

### 1. Configuration réseau (`services/k8s-service/internal/config/config.go`)

Ajout de nouvelles variables d'environnement :
- `K8S_API_TIMEOUT` : Timeout pour les requêtes API (défaut: 30s)
- `K8S_MAX_RETRIES` : Nombre de tentatives en cas d'échec (défaut: 3)

### 2. Validation des endpoints (`services/k8s-service/internal/k8s/client.go`)

- **En production** : Rejet automatique des endpoints locaux
- **En développement** : Autorisation des endpoints locaux avec TLS désactivé
- Validation stricte des certificats TLS en production

### 3. Validation dans les handlers (`services/k8s-service/internal/handler/cluster_handler.go`)

- Validation lors de la création de cluster (`CreateCluster`)
- Validation lors de la mise à jour de cluster (`UpdateCluster`)
- Messages d'erreur clairs pour l'utilisateur

### 4. Test de connexion amélioré (`services/k8s-service/internal/service/cluster_service.go`)

- Timeout configurable pour les tests de connexion
- Retry avec backoff exponentiel
- Validation TLS stricte en production
- Gestion des erreurs réseau améliorée

## Configuration recommandée pour la production

### Variables d'environnement

```bash
ENV=production
K8S_API_TIMEOUT=60s  # Plus long pour les connexions distantes
K8S_MAX_RETRIES=5    # Plus de tentatives pour la résilience
K8S_INCLUSTER=false  # Utiliser kubeconfigs externes
```

### Architecture réseau recommandée

1. **VPN ou réseau privé** : Connecter Kura aux clusters via réseau privé
2. **Load balancer** : Placer un LB devant les API servers Kubernetes
3. **DNS interne** : Utiliser des noms DNS internes pour les clusters
4. **Firewall** : Ouvrir les ports nécessaires (6443 pour API, 10250 pour kubelet)
5. **Monitoring** : Surveiller la connectivité réseau et les latences

### Checklist de déploiement

- [ ] Vérifier que tous les kubeconfigs pointent vers des endpoints accessibles
- [ ] Configurer les certificats TLS valides
- [ ] Tester la connectivité réseau depuis le service vers les clusters
- [ ] Configurer les timeouts appropriés selon la latence réseau
- [ ] Mettre en place un monitoring de la connectivité
- [ ] Documenter les endpoints et la configuration réseau

## Exemples d'erreurs en production

### Endpoint local rejeté
```
{
  "error": "endpoints locaux (127.0.0.1, localhost, host.docker.internal) interdits en production pour des raisons de sécurité"
}
```

### Connexion échouée après retry
```
{
  "connected": false,
  "error": "impossible de se connecter au cluster après 5 tentatives: context deadline exceeded"
}
```

## Notes importantes

- Les endpoints locaux sont **automatiquement rejetés** en production
- La validation TLS est **toujours activée** en production
- Les timeouts peuvent être ajustés selon votre infrastructure
- Le retry automatique améliore la résilience face aux erreurs réseau temporaires
