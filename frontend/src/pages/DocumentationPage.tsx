import { useState, useEffect } from 'react'
import {
  Box,
  Drawer,
  List,
  ListItemButton,
  ListItemText,
  Typography,
  useTheme,
  useMediaQuery,
  IconButton,
  Paper,
} from '@mui/material'
import { Menu as MenuIcon } from '@mui/icons-material'
import { useSearchParams } from 'react-router-dom'
import ModuleTitle from '../components/ModuleTitle'
import CodeBlock from '../components/CodeBlock'
import { jellyfishColors } from '../theme'

const docSidebarWidth = 280

const docSections: { id: string; label: string; children?: { id: string; label: string }[] }[] = [
  { id: 'intro', label: 'Présentation' },
  { id: 'getting-started', label: 'Démarrage rapide' },
  {
    id: 'modules',
    label: 'Modules',
    children: [
      { id: 'k8s', label: 'Kubernetes' },
      { id: 'argocd', label: 'ArgoCD & Helm' },
      { id: 'registry', label: 'Zot' },
      { id: 'terraform', label: 'Terraform' },
      { id: 'code', label: 'Repository' },
      { id: 'ansible', label: 'Ansible' },
      { id: 'vault', label: 'Vault' },
      { id: 'pipelines', label: 'Pipelines CI/CD' },
      { id: 'monitoring', label: 'Observabilité' },
    ],
  },
  {
    id: 'compte',
    label: 'Compte & Projets',
    children: [
      { id: 'users', label: 'Utilisateurs et rôles' },
      { id: 'projects', label: 'Gérer ses projets' },
    ],
  },
  { id: 'faq', label: 'FAQ & Dépannage' },
]

const contentSx = {
  maxWidth: 720,
  mx: 'auto',
  '& h1': { fontSize: '1.75rem', fontWeight: 700, color: jellyfishColors.cyanSoft, mb: 2, mt: 0, lineHeight: 1.3 },
  '& h2': { fontSize: '1.25rem', fontWeight: 600, color: jellyfishColors.cyanLight, mb: 1.5, mt: 3.5, lineHeight: 1.35 },
  '& h3': { fontSize: '1.05rem', fontWeight: 600, color: jellyfishColors.violetSoft, mb: 1, mt: 2, lineHeight: 1.4 },
  '& p': { lineHeight: 1.9, mb: 1.5, color: 'rgba(255,255,255,0.92)' },
  '& ul': { pl: 2.5, mb: 2, '& li': { mb: 0.85, lineHeight: 1.75 } },
  '& code': {
    fontFamily: '"JetBrains Mono", Consolas, monospace',
    fontSize: '0.9em',
    bgcolor: 'rgba(0,0,0,0.45)',
    color: jellyfishColors.cyanLight,
    px: 0.75,
    py: 0.25,
    borderRadius: 0.5,
    border: '1px solid rgba(79,142,247,0.12)',
  },
  '& .code': {
    fontFamily: '"JetBrains Mono", Consolas, monospace',
    fontSize: '0.875rem',
    bgcolor: '#1e1e1e',
    border: '1px solid rgba(255,255,255,0.12)',
    borderRadius: 1,
    p: 1.5,
    mb: 2,
    overflow: 'auto',
    color: 'rgba(255,255,255,0.95)',
  },
  '& .card': {
    borderLeft: `4px solid ${jellyfishColors.cyanSoft}`,
    bgcolor: 'rgba(79,142,247,0.06)',
    borderRadius: 1,
    p: 2,
    mb: 2,
  },
}

function DocContent({ docId }: { docId: string }) {
  switch (docId) {
    case 'intro':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Bienvenue dans Kura</Typography>
          <Typography>
            Kura est une plateforme DevOps unifiée qui vise à donner une vue centrale sur plusieurs briques déjà très utilisées en entreprise : Kubernetes, Terraform, Ansible, pipelines CI/CD, métriques et alertes.
          </Typography>
          <Typography component="h2">Ce qui existe déjà sur le marché</Typography>
          <Typography>
            Les outils spécialisés (portails Kubernetes, Terraform Cloud, Ansible Tower, GitHub Actions, Prometheus/Grafana) sont souvent isolés : les équipes jonglent entre plusieurs interfaces, et il est difficile d&apos;avoir une vision transversale (par exemple : « ce pipeline déploie quels clusters ou quelles ressources Terraform ? »). L&apos;authentification et les rôles sont souvent dupliqués dans chaque outil.
          </Typography>
          <Typography component="h2">La valeur ajoutée de Kura</Typography>
          <Box component="ul">
            <li><strong>Point d&apos;entrée unique</strong> pour les équipes Ops / DevOps : un seul portail, une seule API Gateway.</li>
            <li><strong>Agrégation</strong> des infos clés (clusters, états Terraform, jobs Ansible, pipelines, métriques) au même endroit.</li>
            <li><strong>Authentification centralisée</strong> (auth-service) et rôles homogènes sur tous les modules.</li>
            <li>Focus sur l&apos;<strong>opérationnel actif</strong> (gestion K8s, exécution Terraform, jobs Ansible) plutôt que sur le catalogue seul.</li>
          </Box>
          <Typography component="h2">Ce que vous pouvez faire</Typography>
          <Box component="ul">
            <li>Gérer vos clusters Kubernetes (GKE, AKS, EKS, Proxmox ou générique)</li>
            <li>Consulter et gérer vos états Terraform, détecter les dérives (drift)</li>
            <li>Lancer et suivre des jobs Ansible (via AWX/Tower)</li>
            <li>Suivre les pipelines CI/CD et les webhooks</li>
            <li>Visualiser les métriques et configurer les alertes</li>
          </Box>
          <Typography>
            Toutes les actions sont organisées par <strong>projet</strong>. Créez un projet, sélectionnez-le dans la barre latérale, puis utilisez les modules.
          </Typography>
        </Box>
      )
    case 'getting-started':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Démarrage rapide</Typography>
          <Typography>
            En 3 étapes, vous avez accès à tous vos outils DevOps depuis une interface unique.
          </Typography>
          <Typography component="h2">1. Créer un compte</Typography>
          <Typography>
            Rendez-vous sur la page <strong>Inscription</strong>, renseignez votre email, un nom d&apos;utilisateur et un mot de passe. Vous êtes immédiatement connecté après l&apos;inscription.
          </Typography>
          <Typography component="h2">2. Créer un projet</Typography>
          <Typography>
            Kura organise toutes vos ressources par <strong>projet</strong>. Créez votre premier projet depuis la page <strong>Projets</strong> (icône en bas de la barre latérale). Donnez-lui un nom et une description. Sélectionnez-le dans le menu déroulant en haut à gauche — toutes les actions se déroulent dans le contexte du projet actif.
          </Typography>
          <Typography component="h2">3. Connecter vos outils</Typography>
          <Typography>
            Accédez à la page <strong>Modules</strong> pour voir les briques disponibles. Pour commencer :
          </Typography>
          <Box component="ul">
            <li><strong>Kubernetes</strong> : ajoutez un cluster via son kubeconfig → visualisez pods, services, logs</li>
            <li><strong>Terraform</strong> : uploadez un fichier <code>.tfstate</code> ou liez un bucket cloud → consultez vos ressources et détectez les dérives</li>
            <li><strong>Pipelines</strong> : connectez votre token GitHub → suivez vos workflows en temps réel</li>
            <li><strong>Ansible</strong> : connectez votre instance Semaphore → lancez et suivez vos playbooks</li>
            <li><strong>Vault</strong> : connectez votre instance HashiCorp Vault → parcourez et gérez vos secrets</li>
            <li><strong>Observabilité</strong> : disponible automatiquement, aucune configuration requise</li>
          </Box>
        </Box>
      )
    case 'projects':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Gérer ses projets</Typography>
          <Typography>
            Un <strong>projet</strong> est l&apos;unité d&apos;organisation centrale de Kura. Toutes vos ressources (clusters Kubernetes, états Terraform, connexions Ansible, pipelines) sont rattachées à un projet.
          </Typography>
          <Typography component="h2">Créer un projet</Typography>
          <Typography>
            Depuis la page <strong>Projets</strong> (icône en bas de la barre latérale), cliquez sur <strong>Nouveau projet</strong>. Donnez-lui un nom et optionnellement une description. Une fois créé, sélectionnez-le dans le menu déroulant en haut à gauche pour l&apos;activer.
          </Typography>
          <Typography component="h2">Inviter des collaborateurs</Typography>
          <Typography>
            Depuis la page Projets → votre projet → <strong>Membres</strong>, vous pouvez ajouter des collaborateurs par email. Deux rôles sont disponibles :
          </Typography>
          <Box component="ul">
            <li><strong>Admin</strong> : peut modifier les ressources du projet et gérer les membres</li>
            <li><strong>Member</strong> : accès en lecture et actions sur les ressources, sans gestion des membres</li>
          </Box>
          <Typography component="h2">Changer de projet</Typography>
          <Typography>
            Utilisez le sélecteur de projet en haut de la barre latérale pour basculer entre vos projets. L&apos;ensemble des modules (Kubernetes, Terraform, Ansible…) se met à jour automatiquement avec les ressources du projet sélectionné.
          </Typography>
          <Typography component="h2">Isolation des données</Typography>
          <Typography>
            Chaque projet est isolé : un membre d&apos;un projet ne peut pas voir les ressources d&apos;un autre projet auquel il n&apos;appartient pas. Vous pouvez créer autant de projets que nécessaire pour séparer vos environnements (production, staging, dev) ou vos équipes.
          </Typography>
        </Box>
      )
    case 'k8s':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Kubernetes</Typography>
          <Typography>
            Le module Kubernetes permet de gérer plusieurs clusters (GKE, AKS, EKS, Proxmox ou générique), d&apos;activer un cluster à la fois, puis de consulter les namespaces, pods, deployments, services, ConfigMaps, Secrets et Nodes de ce cluster.
          </Typography>
          <Typography component="h2">Fonctionnalités principales</Typography>
          <Box component="ul">
            <li><strong>Clusters</strong> : ajout d&apos;un cluster (nom, type, kubeconfig ou credentials cloud). Pour GKE, vous pouvez fournir une clé JSON de compte de service GCP. Un seul cluster peut être actif par projet à la fois.</li>
            <li><strong>Ressources</strong> : liste des pods, deployments, services, ConfigMaps, Secrets, Nodes. En cliquant sur une ligne, une modale détail s&apos;ouvre avec les onglets Overview, YAML, et selon le type : Logs, Terminal (pods), Events.</li>
            <li><strong>Pods multi-containers</strong> : pour les logs et le terminal, un sélecteur de container permet de choisir le container à utiliser (évite l&apos;erreur « a container name must be specified »).</li>
            <li><strong>Deployments</strong> : l&apos;onglet Overview affiche les replicas (souhaitées, prêtes, disponibles), le selector, etc. Scale et suppression sont disponibles selon les droits.</li>
          </Box>
          <Typography component="h2">Types de clusters</Typography>
          <Typography>
            Support des clusters <strong>GKE</strong>, <strong>AKS</strong>, <strong>EKS</strong>, <strong>Proxmox</strong> ou <strong>générique</strong>. Pour GKE, en environnement Docker, définir <code>GOOGLE_APPLICATION_CREDENTIALS</code> et monter le fichier de clé JSON si besoin. Voir la section « Connexion GKE » pour le détail.
          </Typography>
        </Box>
      )
    case 'argocd':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module ArgoCD &amp; Catalogue Helm</Typography>
          <Typography>
            Le module ArgoCD apporte le <strong>GitOps</strong> à Kura : ArgoCD surveille en continu un dépôt Git (ou un chart Helm) et synchronise automatiquement le cluster Kubernetes actif avec l&apos;état déclaré. Depuis la page ArgoCD, vous pouvez installer ArgoCD sur le cluster actif, créer des <strong>Applications</strong>, suivre leur statut de synchronisation/santé, déclencher une synchronisation manuelle, consulter l&apos;historique des déploiements et effectuer un rollback.
          </Typography>
          <Typography component="h2">Installation</Typography>
          <Typography>
            Le bouton <strong>Installer ArgoCD</strong> déploie ArgoCD dans le namespace <code>argocd</code> du cluster actif. Une fois le serveur prêt, la section « Applications » devient disponible.
          </Typography>
          <Typography component="h2">Créer une Application — deux sources possibles</Typography>
          <Typography>
            Une Application ArgoCD décrit <strong>quoi déployer</strong> (source) et <strong>où</strong> (destination : cluster + namespace), avec une politique de synchronisation (manuelle ou automatique, avec <em>prune</em> et <em>self-heal</em>). Kura propose deux façons de définir la source :
          </Typography>
          <Box component="ul">
            <li><strong>Nouvelle Application (source Git)</strong> : mode classique/avancé. Vous renseignez l&apos;URL d&apos;un dépôt Git contenant des manifests Kubernetes ou un chart Helm (<code>spec.source.repoURL</code> + <code>spec.source.path</code>), ainsi que la révision cible (branche, tag ou commit).</li>
            <li><strong>Depuis le catalogue Helm (source Helm/ArtifactHub)</strong> : mode guidé pour déployer un outil connu (Prometheus, cert-manager, ingress-nginx, Loki, etc.) sans avoir à gérer un dépôt Git. Le formulaire est pré-rempli automatiquement à partir du chart choisi dans le catalogue.</li>
          </Box>
          <Typography component="h2">Catalogue Helm (ArtifactHub)</Typography>
          <Typography>
            Le bouton <strong>Depuis le catalogue Helm</strong> ouvre une recherche en direct dans <a href="https://artifacthub.io" target="_blank" rel="noreferrer">ArtifactHub</a>, l&apos;annuaire public officiel des charts Helm (charts <strong>CNCF</strong> et <strong>officiels</strong> mis en avant par défaut). Pour chaque chart, Kura affiche le logo, la description, le dépôt source, la version la plus récente et les badges « CNCF » / « Officiel ».
          </Typography>
          <Typography>
            En cliquant sur un chart, le formulaire de création d&apos;Application se pré-remplit avec : l&apos;URL du dépôt Helm, le nom du chart et sa version. Il ne reste plus qu&apos;à renseigner le <strong>nom de l&apos;Application</strong> et le <strong>namespace de destination</strong>, et éventuellement personnaliser les <strong>Values (YAML)</strong> pour surcharger les valeurs par défaut du chart.
          </Typography>
          <Typography component="h3">Pourquoi pas un module Helm séparé ?</Typography>
          <Typography>
            Pour éviter toute <strong>dérive GitOps</strong> (« drift »), Kura ne propose pas d&apos;installer un chart Helm directement (<code>helm install</code>) en dehors d&apos;ArgoCD : cela créerait deux sources de vérité différentes pour le même cluster. Les charts du catalogue sont donc déployés exclusivement via <code>spec.source.helm</code> d&apos;ArgoCD — ArgoCD reste la seule autorité sur l&apos;état du cluster, qu&apos;il s&apos;agisse de manifests Git ou de charts Helm.
          </Typography>
          <Typography component="h2">Suivi et opérations</Typography>
          <Box component="ul">
            <li><strong>Statut de synchronisation</strong> : <code>Synced</code> (conforme), <code>OutOfSync</code> (dérive détectée) ou inconnu.</li>
            <li><strong>Statut de santé</strong> : état de santé Kubernetes de l&apos;Application (Healthy, Progressing, Degraded...).</li>
            <li><strong>Synchroniser</strong> : déclenche manuellement une synchronisation (avec option <em>prune</em> pour supprimer les ressources qui ne sont plus déclarées).</li>
            <li><strong>Rafraîchir</strong> : force ArgoCD à recomparer l&apos;état désiré avec l&apos;état réel du cluster.</li>
            <li><strong>Historique &amp; rollback</strong> : chaque synchronisation est enregistrée ; vous pouvez revenir à une révision précédente directement depuis l&apos;interface.</li>
            <li><strong>Synchronisation automatique</strong> : si activée (avec <em>self-heal</em>), ArgoCD réapplique automatiquement l&apos;état déclaré dès qu&apos;une dérive est détectée, sans action manuelle.</li>
          </Box>
        </Box>
      )
    case 'registry':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Zot</Typography>
          <Typography>
            Le module Zot fournit un <strong>registre OCI privé</strong> pour héberger vos propres images de conteneurs et vos charts Helm, avec un suivi de signature <strong>Cosign</strong>. Il est propulsé par <a href="https://zotregistry.dev" target="_blank" rel="noreferrer">Zot</a>, un registre OCI léger (CNCF Sandbox).
          </Typography>
          <Typography>
            Conformément au principe <strong>« Kura arrive nu »</strong>, Zot n&apos;est pas hébergé sur l&apos;infrastructure de Kura : il est déployé <strong>dans le cluster du client</strong>, au même titre qu&apos;ArgoCD ou cert-manager. Kura s&apos;y connecte à distance via le kubeconfig du cluster actif, en ouvrant un tunnel (port-forward) vers le pod Zot — aucune donnée du client n&apos;est stockée côté Kura.
          </Typography>
          <Typography component="h2">Déployer Zot dans votre cluster</Typography>
          <Typography>
            Zot se déploie comme une <strong>Application ArgoCD</strong> (chart Helm officiel ou manifests Git), dans un namespace dédié <code>zot</code>. Pour que le module Zot de Kura fonctionne, le déploiement doit respecter :
          </Typography>
          <Box component="ul">
            <li>Namespace : <code>zot</code></li>
            <li>Label du pod : <code>app=zot</code> (utilisé par Kura pour localiser le pod via le port-forward)</li>
            <li>Service nommé <code>zot</code>, exposant le port <code>5000</code></li>
          </Box>
          <Typography>
            Une fois déployé et <strong>Synced/Healthy</strong> dans ArgoCD, la page Zot de Kura affiche automatiquement le contenu du registre. Si aucun pod Zot n&apos;est trouvé dans le cluster actif, la page affiche un message d&apos;erreur explicite.
          </Typography>
          <Typography component="h2">Pourquoi Zot plutôt que Harbor ?</Typography>
          <Typography>
            Zot est un binaire Go unique, sans base de données externe ni scanner de vulnérabilités imposé. Harbor impose une stack plus lourde (PostgreSQL, Redis, Trivy). Kura n&apos;intègre <strong>volontairement aucun scanner de vulnérabilités type Trivy</strong> — le module Registre se concentre sur l&apos;hébergement et la <strong>signature cryptographique</strong> des artefacts via Cosign, qui constitue une garantie d&apos;intégrité et de provenance indépendante d&apos;un scanner CVE.
          </Typography>
          <Typography component="h2">Parcourir le registre</Typography>
          <Typography>
            La page Zot liste tous les <strong>dépôts</strong> (repositories) présents dans Zot, avec leur nombre de tags. En cliquant sur un dépôt, vous accédez au détail de chaque tag : son <strong>type</strong> (image de conteneur ou chart Helm, détecté via le <code>mediaType</code> OCI du manifeste), sa <strong>taille</strong>, et son statut de signature.
          </Typography>
          <Typography component="h2">Statut de signature Cosign</Typography>
          <Typography>
            Cosign stocke chaque signature comme un artefact OCI séparé, associé à l&apos;image via un tag au format <code>&lt;algo&gt;-&lt;digest&gt;.sig</code> (ex : <code>sha256-abcd1234....sig</code>). Kura détecte la présence de ce tag pour chaque manifeste et affiche le badge <strong>« Signé (Cosign) »</strong> (vert) ou <strong>« Non signé »</strong> en conséquence — sans appel à un service externe.
          </Typography>
          <Typography component="h2">Pousser des images depuis vos pipelines CI/CD</Typography>
          <Typography>
            Le registre Zot est joignable depuis l&apos;intérieur du cluster client à l&apos;adresse <code>zot.zot.svc.cluster.local:5000</code> (DNS du Service Kubernetes dans le namespace <code>zot</code>). Pour pousser depuis un pipeline CI/CD externe au cluster, exposez le Service via un <code>Ingress</code> ou un <code>LoadBalancer</code> et utilisez cette adresse à la place. Stockez les identifiants éventuels dans <strong>Vault</strong> (voir la section <em>Vault</em>) et utilisez-les dans votre pipeline, par exemple avec Docker :
          </Typography>
          <Box component="ul">
            <li><code>docker login zot.zot.svc.cluster.local:5000 -u $REGISTRY_USER -p $REGISTRY_PASSWORD</code></li>
            <li><code>docker build -t zot.zot.svc.cluster.local:5000/mon-app:1.0.0 .</code></li>
            <li><code>docker push zot.zot.svc.cluster.local:5000/mon-app:1.0.0</code></li>
          </Box>
          <Typography>
            Pour signer l&apos;image après le push avec <strong>Cosign</strong> :
          </Typography>
          <Box component="ul">
            <li><code>cosign generate-key-pair</code> (une seule fois, à conserver dans Vault)</li>
            <li><code>cosign sign --key cosign.key zot.zot.svc.cluster.local:5000/mon-app:1.0.0</code></li>
          </Box>
          <Typography>
            Une fois la signature poussée, le badge « Signé (Cosign) » apparaît automatiquement dans la page Zot pour ce tag.
          </Typography>
          <Typography component="h2">Charts Helm internes et catalogue ArgoCD</Typography>
          <Typography>
            Les charts Helm publiés au format OCI dans Zot (<code>mediaType</code> <code>application/vnd.cncf.helm.config.v1+json</code>) apparaissent automatiquement dans l&apos;onglet <strong>« Zot »</strong> du catalogue Helm de la page ArgoCD, aux côtés des charts publics d&apos;ArtifactHub. La sélection d&apos;un chart interne pré-remplit le formulaire de création d&apos;Application avec <code>oci://zot.zot.svc.cluster.local:5000</code> comme dépôt, exactement comme pour un chart ArtifactHub.
          </Typography>
        </Box>
      )
    case 'terraform':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Terraform</Typography>
          <Typography>
            Le module Terraform permet d&apos;uploader et gérer des états Terraform (tfstate), de les associer à des sources de stockage (S3, GCS, Azure Blob), et de lancer une <strong>détection de drift</strong> entre l&apos;état déclaré dans le tfstate et l&apos;infrastructure réelle (GCP, AWS, Azure selon le provider).
          </Typography>
          <Typography component="h2">Exemple de configuration</Typography>
          <Typography>
            Voici un exemple typique de configuration Terraform (bloc <code>terraform</code> et provider Google) :
          </Typography>
          <CodeBlock language="hcl" label="Bloc Terraform">
            {`terraform {
  required_version = ">= 1.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
}`}
          </CodeBlock>
          <Typography component="h3">Ressource réseau (exemple)</Typography>
          <CodeBlock language="hcl" label="Ressource VPC">
            {`resource "google_compute_network" "vpc" {
  name                    = "mon-vpc"
  auto_create_subnetworks = false
  routing_mode            = "GLOBAL"
  project                 = var.project_id
}`}
          </CodeBlock>
          <Typography component="h2">États et sources</Typography>
          <Typography>
            Vous uploadez un fichier tfstate ou vous le synchronisez depuis une source (bucket S3, GCS, Azure Blob). Chaque état est rattaché à un projet. Les credentials (clés cloud) sont configurées au niveau du service ou du projet selon l&apos;implémentation.
          </Typography>
          <Typography component="h2">Détection de drift</Typography>
          <Typography>
            La détection de drift compare l&apos;état Terraform avec l&apos;infrastructure réelle via les APIs des providers. Pour <strong>GCP</strong>, le service s&apos;appuie sur les APIs Compute, Container (GKE), etc. pour les types courants (réseau, sous-réseau, cluster GKE, node pool), et sur l&apos;API <strong>Cloud Asset Inventory</strong> pour tous les autres types de ressources. Les résultats indiquent pour chaque ressource : <strong>in_sync</strong> (synchronisée), <strong>drifted</strong> (différences détectées), <strong>missing</strong> (ressource absente côté cloud) ou <strong>unknown</strong> (erreur ou type non reconnu). Vous pouvez lancer un drift manuellement depuis l&apos;interface ou après une synchronisation.
          </Typography>
          <Typography component="h2">Résultats</Typography>
          <Typography>
            La page de résultat liste chaque ressource du tfstate avec son statut et un message explicatif. Les différences (attributs modifiés) sont affichées pour les ressources en drift. La note en bas de page rappelle que la détection utilise les APIs réelles des providers pour comparer l&apos;état réel au tfstate.
          </Typography>

          <Typography component="h2">Détection de drift « fine » (via OpenTofu)</Typography>
          <Typography>
            En complément du mode <strong>Fast</strong> (détecteurs dédiés par type de ressource, décrit ci-dessus), Kura propose un mode <strong>Fine</strong> : il exécute <code>tofu plan -refresh-only</code> directement sur vos fichiers <code>.tf</code> récupérés depuis GitHub. Ce mode fonctionne pour <strong>n&apos;importe quel type de ressource</strong>, sur n&apos;importe quel provider (GCP, AWS, Azure…), sans nécessiter de support spécifique côté Kura.
          </Typography>
          <Typography component="h3">1. Configurer un token GitHub</Typography>
          <Typography>
            Depuis l&apos;onglet <strong>Configuration</strong> du module Terraform, renseignez un <strong>token d&apos;accès personnel GitHub</strong> (scope <code>repo</code> pour un dépôt privé, ou <code>public_repo</code> pour un dépôt public). Un badge indique si un token est configuré (« Token configuré » / « Aucun token configuré »). Le token est partagé par toutes les sources du projet et peut être supprimé via le bouton <strong>Supprimer</strong>.
          </Typography>
          <Typography component="h3">2. Configurer le dépôt sur la source</Typography>
          <Typography>
            Dans le formulaire d&apos;ajout/édition d&apos;une source (section « Drift fine (optionnel) »), renseignez :
          </Typography>
          <Box component="ul">
            <li><strong>Owner</strong> : le compte ou l&apos;organisation GitHub (ex : <code>MPFabio</code> pour <code>github.com/MPFabio/mon-repo</code>)</li>
            <li><strong>Repo</strong> : le nom du dépôt (ex : <code>mon-repo</code>)</li>
            <li><strong>Path</strong> : le chemin du dossier contenant les fichiers <code>.tf</code> (ex : <code>terraform</code>, ou vide pour la racine)</li>
            <li><strong>Ref</strong> : la branche ou le tag à utiliser (par défaut <code>main</code>)</li>
          </Box>
          <Typography component="h3">3. Conventions attendues pour vos fichiers .tf</Typography>
          <Box component="ul">
            <li>Seuls les fichiers <code>.tf</code> situés <strong>directement</strong> dans le dossier indiqué par <em>Path</em> sont chargés (pas les sous-dossiers/modules).</li>
            <li>Tout bloc <code>backend &quot;...&quot; {'{ }'}</code> à l&apos;intérieur du bloc <code>terraform {'{ }'}</code> est automatiquement retiré : Kura utilise toujours le tfstate déjà importé dans la plateforme, jamais votre backend distant (S3, GCS, etc.).</li>
            <li>
              Les <strong>variables sans valeur par défaut</strong> sont résolues automatiquement, dans l&apos;ordre suivant :
              <Box component="ol" sx={{ pl: 2.5, mt: 0.5, '& li': { mb: 0.5 } }}>
                <li>une <em>output</em> du tfstate portant exactement le même nom ;</li>
                <li>pour les variables contenant <code>zone</code> ou <code>region</code> : une output dont le nom contient ce mot-clé, ou une région dérivée d&apos;une output « zone » (ex : <code>europe-west1-b</code> → <code>europe-west1</code>) ;</li>
                <li>une valeur par défaut neutre selon le type déclaré (<code>0</code> pour <code>number</code>, <code>false</code> pour <code>bool</code>, chaîne vide sinon).</li>
              </Box>
            </li>
            <li>
              Ces heuristiques évitent d&apos;avoir à modifier votre dépôt, mais peuvent produire des valeurs imprécises si aucune output pertinente n&apos;existe dans le tfstate (ex : <code>node_count</code> résolu à <code>0</code>). Pour des résultats fiables, déclarez des <code>default</code> explicites dans <code>variables.tf</code> pour les variables qui ne correspondent à aucune output.
            </li>
          </Box>
          <Typography component="h3">4. Lancer une détection et lire le résultat</Typography>
          <Typography>
            Le paramètre <code>method</code> contrôle le mode utilisé : <code>fine</code> (force le mode fin, erreur si aucun dépôt n&apos;est configuré sur la source), <code>fast</code> (détecteurs existants), ou <code>auto</code> (par défaut : <strong>fine</strong> si un dépôt GitHub est configuré sur la source, sinon <strong>fast</strong>, avec repli automatique sur <strong>fast</strong> si le mode fine échoue).
          </Typography>
          <Typography>
            Chaque résultat de drift affiche un badge <strong>Fine</strong> ou <strong>Fast</strong> selon la méthode utilisée pour l&apos;obtenir. En mode <strong>Fine</strong>, les statuts <strong>drifted</strong>/<strong>in_sync</strong>/<strong>missing</strong> reflètent le diff <code>tofu plan -refresh-only</code> attribut par attribut (chemin pointé, valeur attendue vs. valeur réelle), pour n&apos;importe quel type de ressource déclaré dans vos fichiers <code>.tf</code>.
          </Typography>
          <Typography component="h3">Bruit lié à l&apos;évolution du schéma du provider</Typography>
          <Typography>
            Lorsqu&apos;une nouvelle version du provider (ex : <code>hashicorp/google</code>) introduit des attributs ou blocs calculés qui n&apos;existaient pas au moment de l&apos;écriture du tfstate, ces blocs peuvent techniquement apparaître comme des différences (ex : <code>node_config[0]</code>, <code>node_pool[0]</code>, <code>null</code> → bloc renseigné) même si l&apos;infrastructure n&apos;a pas changé. Kura filtre automatiquement ces écarts (valeurs <code>null</code> ↔ vides, et ajout d&apos;un bloc complet là où l&apos;ancien state avait <code>null</code>) afin de ne conserver que les véritables différences sur des attributs scalaires existants (ex : <code>master_version</code>).
          </Typography>
        </Box>
      )
    case 'code':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Repository</Typography>
          <Typography>
            Le module <strong>Repository</strong> permet de parcourir, en lecture seule, le code source des dépôts GitHub liés à votre projet : arborescence des fichiers, contenu avec coloration syntaxique, et historique des commits avec leurs diffs.
          </Typography>
          <Typography component="h2">1. Lier un ou plusieurs dépôts au projet</Typography>
          <Typography>
            Un projet peut être lié à <strong>plusieurs dépôts GitHub</strong> via les mappings de projet (page Projets → votre projet → dépôts liés). Chaque dépôt lié apparaît dans le sélecteur de dépôt en haut de la page Code.
          </Typography>
          <Typography component="h2">2. Configurer le token GitHub</Typography>
          <Typography>
            Le module Repository réutilise le <strong>token GitHub</strong> déjà configuré pour la détection de drift Terraform (Terraform → Configuration). Aucune configuration supplémentaire n&apos;est nécessaire si ce token est déjà renseigné et dispose d&apos;un accès en lecture sur les dépôts liés.
          </Typography>
          <Typography component="h2">3. Naviguer dans les fichiers</Typography>
          <Typography>
            L&apos;onglet <strong>Fichiers</strong> affiche l&apos;arborescence du dépôt sélectionné (à la branche/ref indiquée, <code>main</code> par défaut). Cliquez sur un dossier pour le déplier, ou sur un fichier pour afficher son contenu avec coloration syntaxique. Les fichiers de plus de 1 Mo ne sont pas affichés (aperçu indisponible).
          </Typography>
          <Typography component="h2">4. Consulter l&apos;historique des commits</Typography>
          <Typography>
            L&apos;onglet <strong>Historique</strong> liste les derniers commits de la branche sélectionnée. Cliquez sur un commit pour afficher le détail des fichiers modifiés et leur diff (lignes ajoutées en vert, supprimées en rouge).
          </Typography>
        </Box>
      )
    case 'ansible':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Ansible</Typography>
          <Typography>
            Le module Ansible vous permet de suivre et lancer vos playbooks Ansible directement depuis Kura, sans quitter l&apos;interface.
          </Typography>
          <Typography component="h2">Connecter votre instance Ansible</Typography>
          <Typography>
            Kura fonctionne avec <strong>Ansible Semaphore</strong> (interface open-source pour Ansible). Depuis la page Ansible, cliquez sur <strong>Connecter un backend Ansible</strong> et renseignez :
          </Typography>
          <Box component="ul">
            <li><strong>URL Semaphore</strong> : l&apos;adresse de votre instance Semaphore</li>
            <li><strong>Token API</strong> : généré dans Semaphore → User Settings → API Tokens → New Token</li>
            <li><strong>Project ID</strong> : visible dans l&apos;URL de votre projet (<code>/project/<strong>1</strong>/</code>)</li>
          </Box>
          <Typography>
            Cliquez <strong>Connecter</strong>. Le badge passe en vert dès que la connexion est établie. La configuration est mémorisée.
          </Typography>
          <Typography component="h2">Ce que vous pouvez faire</Typography>
          <Box component="ul">
            <li><strong>Jobs</strong> : voir les exécutions en cours (statut, durée, template utilisé)</li>
            <li><strong>Historique</strong> : consulter toutes les exécutions passées et leurs résultats</li>
            <li><strong>Inventaires</strong> : voir les groupes d&apos;hôtes configurés dans Semaphore</li>
            <li><strong>Templates</strong> : lister les playbooks disponibles et en lancer un en un clic (bouton ▶)</li>
          </Box>
          <Typography component="h2">Lancer un playbook</Typography>
          <Typography>
            Allez dans l&apos;onglet <strong>Templates</strong>, repérez le template souhaité et cliquez sur ▶. L&apos;exécution apparaît immédiatement dans l&apos;onglet <strong>Jobs</strong> avec son statut en temps réel.
          </Typography>
        </Box>
      )
    case 'vault':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Vault</Typography>
          <Typography>
            Le module Vault vous permet de parcourir, créer et supprimer des secrets stockés dans votre instance <strong>HashiCorp Vault</strong>, directement depuis Kura.
          </Typography>
          <Typography component="h2">Connecter votre instance Vault</Typography>
          <Typography>
            Kura ne fournit pas de Vault : vous connectez <strong>votre propre instance</strong> (auto-hébergée, ou HCP Vault). Depuis la page Vault, ouvrez le panneau <strong>Connexion Vault</strong> et renseignez :
          </Typography>
          <Box component="ul">
            <li><strong>Adresse Vault</strong> : l&apos;URL de votre instance Vault, joignable depuis Kura (ex : <code>https://vault.monentreprise.com:8200</code>)</li>
            <li><strong>Token Vault</strong> : un token avec les droits de lecture/écriture sur le mount KV utilisé (créé via <code>vault token create</code> ou une policy dédiée)</li>
            <li><strong>Mount path</strong> : le chemin du moteur KV v2 à utiliser (par défaut <code>secret</code>)</li>
          </Box>
          <Typography>
            Cliquez <strong>Connecter</strong>. Le badge passe à <strong>Connecté</strong> et indique l&apos;état de scellement (<strong>Scellé</strong> / <strong>Déscellé</strong>) de votre Vault. La configuration est mémorisée par projet.
          </Typography>
          <Typography component="h2">Ce que vous pouvez faire</Typography>
          <Box component="ul">
            <li><strong>Parcourir</strong> : naviguer dans l&apos;arborescence des secrets (dossiers et clés) via le fil d&apos;Ariane</li>
            <li><strong>Consulter</strong> : voir les clés d&apos;un secret, révéler ou masquer chaque valeur, et la copier dans le presse-papier</li>
            <li><strong>Créer</strong> : ajouter un nouveau secret à un chemin donné, avec une ou plusieurs paires clé/valeur</li>
            <li><strong>Supprimer</strong> : retirer un secret de Vault après confirmation</li>
          </Box>
          <Typography component="h2">Sécurité</Typography>
          <Typography>
            Le token Vault est stocké côté serveur et n&apos;est jamais ré-affiché en clair (la configuration renvoie <code>***</code>). Les valeurs des secrets ne sont révélées que dans l&apos;UI, à la demande, et transitent toujours via une connexion authentifiée à Kura.
          </Typography>
        </Box>
      )
    case 'monitoring':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Observabilité</Typography>
          <Typography>
            Le module Observabilité réunit les trois piliers — <strong>métriques</strong> (Prometheus), <strong>logs</strong> (Loki) et <strong>traces</strong> (Tempo) — pour l&apos;ensemble des services Kura, avec un dashboard <strong>Grafana</strong> intégré directement dans l&apos;interface. La page est organisée en trois onglets : Métriques, Logs et Traces.
          </Typography>
          <Typography component="h2">Onglet Métriques</Typography>
          <Box component="ul">
            <li><strong>KPI globaux</strong> : nombre de services actifs / hors ligne, goroutines totales, mémoire totale</li>
            <li><strong>Health cards</strong> : état UP/DOWN de chaque service (Auth, Kubernetes, Terraform, Ansible, Pipeline, Vault, Code, Metrics) — vérifié par health check direct</li>
            <li><strong>Tableau de métriques</strong> : goroutines, CPU rate, mémoire RSS par service (données Prometheus)</li>
            <li><strong>Dashboard Grafana</strong> : vue temporelle des métriques (goroutines, mémoire, état des services dans le temps)</li>
          </Box>
          <Typography>
            Le <strong>metrics-service</strong> interroge l&apos;API HTTP de Prometheus (<code>/api/v1/query</code>) et effectue des health checks directs sur chaque service. Les données sont mises en cache 30 secondes dans Redis pour éviter de surcharger Prometheus. La page se rafraîchit automatiquement toutes les 30 secondes.
          </Typography>
          <Typography component="h2">Onglet Logs</Typography>
          <Typography>
            Affiche les logs centralisés de tous les services, collectés par <strong>Promtail</strong> et stockés dans <strong>Loki</strong>. Vous pouvez filtrer par service et effectuer une recherche texte sur les lignes de log. La liste se rafraîchit automatiquement toutes les 15 secondes.
          </Typography>
          <Typography component="h2">Onglet Traces</Typography>
          <Typography>
            Affiche les traces distribuées collectées via <strong>OpenTelemetry</strong> et stockées dans <strong>Tempo</strong>. Chaque service est instrumenté (middleware HTTP automatique) et exporte ses spans vers Tempo. Vous pouvez filtrer par service et par durée minimale, et déplier une trace pour voir le détail de ses spans (chronologie, durée de chaque étape).
          </Typography>
          <Typography>
            Les endpoints <code>/health</code> et <code>/metrics</code> (healthchecks Docker et scraping Prometheus) sont exclus du tracing pour éviter de saturer Tempo avec du bruit à très haute fréquence.
          </Typography>
          <Typography component="h2">Dashboard Grafana</Typography>
          <Typography>
            Le dashboard <strong>Kura Platform Overview</strong> est accessible à <code>/grafana</code> (en production) ou <code>http://localhost:3000</code> (en local). Grafana est préconfiguré avec les sources de données Prometheus, Loki, Tempo et Alertmanager. L&apos;accès anonyme est activé (lecture seule) pour permettre l&apos;affichage dans l&apos;iframe Kura.
          </Typography>
        </Box>
      )
    case 'pipelines':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Pipelines</Typography>
          <Typography>
            Le module Pipelines permet de suivre les pipelines CI/CD (GitHub Actions, GitLab CI, Jenkins, etc.) et de gérer les webhooks. Les exécutions et statuts sont agrégés dans la plateforme pour une vue transversale avec les autres modules (Kubernetes, Terraform, Ansible).
          </Typography>

          <Typography component="h2">Option recommandée : Webhooks (temps réel)</Typography>
          <Typography>
            Les webhooks permettent une mise à jour <strong>immédiate</strong> des exécutions dès qu&apos;un workflow GitHub se termine. Aucun polling, aucune latence.
          </Typography>
          <Typography component="h3">Configurer le webhook sur GitHub</Typography>
          <Box component="ol" sx={{ pl: 2.5, mb: 2, '& li': { mb: 1 } }}>
            <li>Dans votre dépôt : <strong>Settings</strong> → <strong>Webhooks</strong> → <strong>Add webhook</strong></li>
            <li><strong>Payload URL</strong> : saisissez l&apos;URL de votre instance Kura. L&apos;URL complète est affichée dans la page Pipelines (section « Connecter un dépôt GitHub » → « Option temps réel »). Exemple : <code>https://votre-domaine/api/v1/pipeline/webhooks/github</code></li>
            <li><strong>Content type</strong> : <code>application/json</code></li>
            <li><strong>Events</strong> : sélectionnez « Let me select individual events » puis cochez <code>Workflow run</code> (GitHub Actions)</li>
          </Box>

          <Typography component="h2">Option alternative : Synchronisation par API</Typography>
          <Typography>
            Si les webhooks ne peuvent pas être utilisés, utilisez la synchronisation par API : créez un Personal Access Token GitHub (scope <code>repo</code> ou <code>actions:read</code>), puis dans la page Pipelines, section « Connecter un dépôt GitHub », collez le token et indiquez les dépôts au format <code>owner/repo</code>. Les runs sont récupérés périodiquement ou manuellement via le bouton « Synchroniser ».
          </Typography>

          <Typography component="h2">Intégration</Typography>
          <Typography>
            Le pipeline-service peut recevoir des webhooks des plateformes CI/CD, enregistrer les exécutions et les afficher dans l&apos;interface.
          </Typography>
        </Box>
      )
    case 'users':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Utilisateurs et rôles</Typography>
          <Typography component="h2">Créer un compte</Typography>
          <Typography>
            L&apos;inscription est accessible depuis la page d&apos;accueil. Renseignez votre email, un nom d&apos;utilisateur et un mot de passe (minimum 8 caractères). Une fois connecté, vous restez authentifié grâce au refresh token — vous n&apos;avez pas à vous reconnecter à chaque session.
          </Typography>
          <Typography component="h2">Changer son mot de passe</Typography>
          <Typography>
            Depuis <strong>Paramètres</strong> (icône en bas de la barre latérale) → <strong>Sécurité</strong> → saisissez votre mot de passe actuel puis le nouveau.
          </Typography>
          <Typography component="h2">Rôles dans un projet</Typography>
          <Box component="ul">
            <li><strong>Propriétaire</strong> : créateur du projet, accès complet incluant la suppression du projet et la gestion des membres</li>
            <li><strong>Admin</strong> : peut modifier les ressources et gérer les membres</li>
            <li><strong>Member</strong> : accès en lecture et actions sur les ressources (pods, terraform, pipelines…)</li>
          </Box>
          <Typography component="h2">Déconnexion</Typography>
          <Typography>
            Cliquez sur votre avatar en haut à gauche → <strong>Déconnexion</strong>. Votre session est invalidée côté serveur.
          </Typography>
        </Box>
      )
    case 'faq':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">FAQ & Dépannage</Typography>

          <Typography component="h2">Le module Ansible affiche « Aucun historique disponible »</Typography>
          <Typography>
            Vérifiez que la connexion Semaphore est configurée : page Ansible → panneau <em>Connecter un backend Ansible</em> → le badge doit être vert. Si le token ou l&apos;ID projet est incorrect, reconfigurez-les. Lancez un job dans Semaphore, puis actualisez.
          </Typography>

          <Typography component="h2">Vault affiche « Erreur lors de la connexion »</Typography>
          <Typography>
            Vérifiez que l&apos;<strong>adresse Vault</strong> est joignable depuis l&apos;infrastructure Kura (pas une adresse <code>localhost</code> ou un réseau privé inaccessible), que le <strong>token</strong> est valide et non expiré, et que le <strong>mount path</strong> correspond bien à un moteur KV v2 activé sur votre Vault (<code>vault secrets list</code>).
          </Typography>

          <Typography component="h2">Kubernetes — « Impossible de récupérer les namespaces »</Typography>
          <Typography>
            Vérifiez que le cluster est bien ajouté et activé (bouton Activer dans la liste des clusters). Pour GKE, le service account utilisé par Kura doit avoir le rôle <code>roles/container.admin</code> dans IAM GCP. Vous pouvez aussi créer un ClusterRoleBinding Kubernetes : <code>kubectl create clusterrolebinding kura-sa-admin --clusterrole=cluster-admin --user=&lt;sa-email&gt;</code>.
          </Typography>

          <Typography component="h2">Le dashboard Grafana affiche « Dashboard not found »</Typography>
          <Typography>
            Vérifiez que le fichier <code>kura-overview.json</code> est bien dans le dossier de provisioning Grafana. Redémarrez Grafana : <code>docker compose restart grafana</code>. En production, le dashboard est accessible à <code>/grafana/d/kura-overview/...</code>.
          </Typography>

          <Typography component="h2">Terraform — « No configuration found » dans Kura</Typography>
          <Typography>
            Le tfstate existe dans GCS mais est vide (terraform apply n&apos;a pas encore été exécuté) ou la source cloud n&apos;est pas correctement configurée. Vérifiez que le <em>Bucket GCP</em> correspond au nom réel du bucket GCS (ex. <code>kura-ynov</code>) et que le <em>chemin de l&apos;objet</em> est le chemin dans le bucket (ex. <code>demo-kura/state/default.tfstate</code>).
          </Typography>

          <Typography component="h2">Pipeline — les runs ne s&apos;affichent pas</Typography>
          <Typography>
            Configurez la connexion GitHub dans la page Pipelines (token + dépôts). Cliquez sur <strong>Synchroniser</strong> pour forcer la récupération immédiate. Les webhooks GitHub permettent une mise à jour en temps réel sans polling.
          </Typography>

          <Typography component="h2">Page blanche au chargement</Typography>
          <Typography>
            Certains bloqueurs de publicité (uBlock Origin, etc.) peuvent bloquer les requêtes vers des domaines <code>.nip.io</code>. Désactivez l&apos;extension pour ce domaine, ou utilisez un navigateur sans extensions lors des démonstrations.
          </Typography>

          <Typography component="h2">Réinitialiser la base de données</Typography>
          <Typography>
            En local : <code>docker compose down -v && docker compose up -d</code>. <strong>Attention</strong> : cela supprime toutes les données (utilisateurs, projets, clusters). En production, évitez — utilisez <code>docker compose down</code> (sans <code>-v</code>) pour conserver les volumes.
          </Typography>
        </Box>
      )
    default:
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Documentation</Typography>
          <Typography>
            Choisissez une section dans le menu à gauche pour afficher son contenu.
          </Typography>
        </Box>
      )
  }
}

export default function DocumentationPage() {
  const [searchParams] = useSearchParams()
  const sectionParam = searchParams.get('section')
  const [selectedId, setSelectedId] = useState(sectionParam || 'intro')
  const [sidebarOpen, setSidebarOpen] = useState(false)

  useEffect(() => {
    if (sectionParam) {
      const validIds = ['intro', 'getting-started', 'k8s', 'argocd', 'registry', 'terraform', 'ansible', 'pipelines', 'monitoring', 'users', 'projects', 'faq']
      if (validIds.includes(sectionParam)) {
        setSelectedId(sectionParam)
      }
    }
  }, [sectionParam])
  const theme = useTheme()
  const isMobile = useMediaQuery(theme.breakpoints.down('md'))

  const sidebar = (
    <Box
      sx={{
        width: docSidebarWidth,
        flexShrink: 0,
        borderRight: `1px solid ${jellyfishColors.cyanSubtle}`,
        bgcolor: 'rgba(0,0,0,0.25)',
        height: '100%',
        overflowY: 'auto',
      }}
    >
      <List dense sx={{ py: 1 }}>
        {docSections.map((section) => (
          <Box key={section.id}>
            {section.children ? (
              <>
                {/* Groupe parent — non cliquable, label seul */}
                <Box sx={{ px: 2, pt: 1.5, pb: 0.5 }}>
                  <ListItemText primary={section.label} primaryTypographyProps={{ fontSize: '0.7rem', fontWeight: 700, color: jellyfishColors.grayLight, textTransform: 'uppercase', letterSpacing: '0.08em' }} />
                </Box>
                {section.children.map((child) => (
                  <ListItemButton
                    key={child.id}
                    selected={selectedId === child.id}
                    onClick={() => {
                      setSelectedId(child.id)
                      if (isMobile) setSidebarOpen(false)
                    }}
                    sx={{
                      pl: 3,
                      py: 0.6,
                      '&.Mui-selected': { bgcolor: 'rgba(79,142,247,0.10)', borderRight: `3px solid ${jellyfishColors.cyanSoft}` },
                    }}
                  >
                    <ListItemText primary={child.label} primaryTypographyProps={{ fontSize: '0.85rem' }} />
                  </ListItemButton>
                ))}
              </>
            ) : (
              <ListItemButton
                selected={selectedId === section.id}
                onClick={() => {
                  setSelectedId(section.id)
                  if (isMobile) setSidebarOpen(false)
                }}
                sx={{
                  py: 0.75,
                  '&.Mui-selected': { bgcolor: 'rgba(79,142,247,0.12)', borderRight: `3px solid ${jellyfishColors.cyanSoft}` },
                }}
              >
                <ListItemText primary={section.label} primaryTypographyProps={{ fontSize: '0.9rem', fontWeight: 600 }} />
              </ListItemButton>
            )}
          </Box>
        ))}
      </List>
    </Box>
  )

  return (
    <Box sx={{ display: 'flex', flexDirection: 'column', minHeight: '100%' }}>
      <Box sx={{ display: 'flex', alignItems: 'center', gap: 2, mb: 3 }}>
        {isMobile && (
          <IconButton onClick={() => setSidebarOpen(true)} sx={{ color: jellyfishColors.cyanSoft }}>
            <MenuIcon />
          </IconButton>
        )}
        <ModuleTitle>Documentation</ModuleTitle>
      </Box>

      <Box sx={{ display: 'flex', flex: 1, minHeight: 0, gap: 0 }}>
        {!isMobile && sidebar}
        {isMobile && (
          <Drawer
            variant="temporary"
            open={sidebarOpen}
            onClose={() => setSidebarOpen(false)}
            ModalProps={{ keepMounted: true }}
            sx={{
              '& .MuiDrawer-paper': {
                width: docSidebarWidth,
                boxSizing: 'border-box',
                bgcolor: 'rgba(26,35,50,0.98)',
                borderRight: `1px solid ${jellyfishColors.cyanSubtle}`,
              },
            }}
          >
            {sidebar}
          </Drawer>
        )}

        <Paper
          component="article"
          elevation={0}
          sx={{
            flex: 1,
            minWidth: 0,
            p: 4,
            pt: 3,
            pb: 6,
            bgcolor: 'rgba(26,35,50,0.75)',
            border: `1px solid ${jellyfishColors.cyanSubtle}`,
            borderRadius: 2,
            overflowY: 'auto',
            boxShadow: `inset 0 1px 0 rgba(79,142,247,0.05)`,
          }}
        >
          <DocContent docId={selectedId} />
        </Paper>
      </Box>
    </Box>
  )
}
