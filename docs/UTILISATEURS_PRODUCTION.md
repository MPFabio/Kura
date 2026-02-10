# Utilisateurs en production – Kura

Ce document décrit ce qu’il en est des **personnes qui utiliseront la solution** une fois celle-ci mise en production et hébergée.

---

## 1. Qui sont les utilisateurs ?

Toute personne disposant d’un **compte** sur la plateforme peut l’utiliser :

- **Inscription** : aujourd’hui, tout le monde peut s’inscrire via la page **Inscription** (`/register`) (email, nom d’utilisateur, mot de passe).
- **Connexion** : chaque utilisateur se connecte via **Connexion** (`/login`) avec email + mot de passe.
- **Rôles** : le backend gère les rôles (ex. `admin`, `user`). Par défaut, un nouvel inscrit a le rôle `user`.

En production, vous pouvez :
- **Garder l’inscription ouverte** : les futurs utilisateurs créent eux‑mêmes leur compte.
- **Réserver la création de comptes aux admins** : désactiver la route publique d’inscription et créer les comptes via l’API ou un outil interne (à prévoir côté auth-service).

---

## 2. Comment les utilisateurs accèdent-ils aux données ?

- **Projets** : chaque utilisateur a une liste de **projets** auxquels il a accès (projets qu’il a créés ou auxquels il a été ajouté comme membre).
- **Projet courant** : l’utilisateur choisit un projet (page Projets). Toutes les actions (Kubernetes, Terraform, Ansible, Pipelines, etc.) sont faites **dans le cadre de ce projet**.
- **Isolation** : les données (clusters K8s, états Terraform, etc.) sont liées à un `project_id`. Un utilisateur ne voit que les projets auxquels il appartient, donc uniquement les ressources de ces projets.

En résumé : **une fois en prod et hébergé, les personnes qui utilisent la solution sont celles qui ont un compte ; elles ne voient que les projets dont elles sont membres.**

---

## 3. Travail en équipe (plusieurs personnes sur un même projet)

Le modèle actuel le permet déjà côté backend :

- Un projet a un **propriétaire** (`owner_id`) et peut avoir des **membres** (rôles `admin` ou `member`).
- L’API auth-service expose :
  - **Ajout de membre** : `POST /api/v1/projects/:id/members` (avec `user_id` et `role`).
  - **Liste des membres** : `GET /api/v1/projects/:id/members`.
- **Contrôle d’accès** : `UserHasAccessToProject(userID, projectID)` — un utilisateur ne peut accéder qu’aux projets où il est owner ou membre.

Donc **plusieurs personnes peuvent utiliser la même solution** en étant ajoutées au même projet. Aujourd’hui, l’interface peut ne pas encore proposer d’écran « Inviter un membre » ; en prod, il est pertinent d’exposer cette fonctionnalité (appels aux routes ci‑dessus) pour que les équipes gèrent elles‑mêmes les accès.

---

## 4. Parcours type pour un utilisateur (prod hébergée)

1. **Accès à la plateforme** : URL de la solution hébergée (ex. `https://kura.votredomaine.com`).
2. **Compte** : inscription (si ouverte) ou réception d’un compte créé par un admin.
3. **Connexion** : email + mot de passe.
4. **Projet** :  
   - soit création d’un nouveau projet (nom, description),  
   - soit sélection d’un projet existant (s’il est déjà membre).
5. **Modules** : utilisation des modules (Kubernetes, Terraform, Ansible, Pipelines, etc.) dans le cadre du projet sélectionné.

Les **personnes qui utiliseront la solution** en prod suivent ce parcours ; aucune donnée d’un autre projet ne leur est visible.

---

## 5. À prévoir côté hébergement / exploitation

- **HTTPS** : obligatoire en production (certificat, reverse proxy / Kong).
- **Base de données (auth-service)** : sauvegarde régulière de PostgreSQL (utilisateurs, projets, membres, tokens).
- **Secrets** : `JWT_SECRET` et secrets des services doivent être gérés de façon sécurisée (variables d’environnement, coffre).
- **Inscription** : décider si l’inscription reste ouverte ou réservée aux admins ; si réservée, prévoir un moyen de création de comptes (API, script, future UI admin).
- **Documentation utilisateur** (optionnel) : courte notice « Première connexion », « Créer un projet », « Inviter un collègue » (dès que l’UI le permet).

---

## 6. Résumé

| Question | Réponse |
|----------|--------|
| **Qui peut utiliser la solution en prod ?** | Toute personne avec un compte (créé par elle‑même ou par un admin). |
| **Comment obtient‑on un compte ?** | Aujourd’hui : inscription sur `/register`. En prod : possible de désactiver l’inscription et de créer les comptes côté admin. |
| **Les utilisateurs voient‑ils les données des autres ?** | Non. Accès limité aux projets dont ils sont owner ou membre. |
| **Plusieurs personnes sur un même projet ?** | Oui. Le backend gère les membres ; exposer « Inviter un membre » dans l’UI est recommandé en prod. |

En résumé : **une fois en prod et hébergé, les personnes qui utiliseront la solution sont les utilisateurs avec un compte ; elles n’ont accès qu’aux projets auxquels elles appartiennent. Le modèle (projets, membres, rôles) est déjà en place côté backend pour gérer équipes et accès.**
