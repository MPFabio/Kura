import { useState } from 'react'
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
      { id: 'terraform', label: 'Terraform' },
      { id: 'ansible', label: 'Ansible' },
      { id: 'pipelines', label: 'Pipelines' },
    ],
  },
  { id: 'architecture', label: 'Architecture' },
  {
    id: 'production',
    label: 'Production',
    children: [
      { id: 'users', label: 'Utilisateurs et accès' },
      { id: 'gke', label: 'Connexion GKE' },
    ],
  },
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
    border: '1px solid rgba(0,229,255,0.15)',
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
    bgcolor: 'rgba(0,229,255,0.06)',
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
            <li><strong>Événements corrélés</strong> via Kafka (déploiement Terraform, métriques, alertes).</li>
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
          <Typography component="h2">1. Créer un compte</Typography>
          <Typography>
            Inscrivez-vous via la page <strong>Inscription</strong> (email, nom d&apos;utilisateur, mot de passe) ou connectez-vous si vous avez déjà un compte. Les mots de passe sont hachés (bcrypt) et les sessions gérées par JWT et refresh tokens.
          </Typography>
          <Typography component="h2">2. Créer ou rejoindre un projet</Typography>
          <Typography>
            Sur la page <strong>Projets</strong>, créez un nouveau projet (nom, description) ou sélectionnez un projet existant si vous y avez accès. Le projet courant est affiché dans le sélecteur en haut de la barre latérale ; toutes les ressources (clusters K8s, états Terraform, etc.) sont rattachées au projet sélectionné.
          </Typography>
          <Typography component="h2">3. Utiliser les modules</Typography>
          <Typography>
            Une fois un projet sélectionné, accédez aux modules depuis le menu : <strong>Kubernetes</strong>, <strong>Terraform</strong>, <strong>Ansible</strong>, <strong>Pipelines</strong>, <strong>Monitoring</strong>, <strong>Alertes</strong>. Chaque module affiche uniquement les ressources liées au projet courant. Vous pouvez créer des clusters, uploader des états Terraform, lancer des jobs Ansible ou consulter les pipelines selon les droits de votre rôle.
          </Typography>
          <Typography component="h2">Conseils</Typography>
          <Typography>
            Pour une première prise en main : créez un projet de test, ajoutez un cluster Kubernetes (ou un état Terraform) depuis les modules correspondants, puis explorez les vues détail (Overview, YAML, Logs, Terminal pour les pods).
          </Typography>
        </Box>
      )
    case 'architecture':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Architecture</Typography>
          <Typography>
            La plateforme repose sur une architecture microservices : un frontend React (TypeScript), une API Gateway Kong, et des services backend (auth, Kubernetes, Terraform, Ansible, Pipelines) qui s&apos;appuient sur PostgreSQL, Redis et Kafka.
          </Typography>
          <Typography component="h2">Vue globale</Typography>
          <Typography>
            Le frontend envoie toutes les requêtes à Kong. Kong route vers le bon service (auth-service, k8s-service, terraform-service, etc.). Les services métier publient et consomment des événements via Kafka pour corréler les actions (déploiements, alertes, métriques). PostgreSQL stocke les utilisateurs, projets, clusters, états Terraform, etc. Redis est utilisé pour le cache (listes de ressources, etc.).
          </Typography>
          <Typography component="h2">Authentification</Typography>
          <Typography>
            L&apos;authentification est centralisée dans le <strong>auth-service</strong> : inscription, connexion, JWT et refresh tokens. Toutes les requêtes protégées passent par Kong avec le header <code>Authorization: Bearer &lt;token&gt;</code>. Le auth-service valide le token et expose l&apos;ID utilisateur aux handlers ; les accès aux projets sont vérifiés via <code>UserHasAccessToProject</code> (propriétaire ou membre du projet).
          </Typography>
          <Typography component="h2">Données par projet</Typography>
          <Typography>
            Clusters K8s, états Terraform, sources, etc. sont rattachés à un <strong>project_id</strong>. Seuls les utilisateurs ayant accès au projet (propriétaire ou membre) peuvent voir et modifier ces données. Le découpage par domaines (auth-service, k8s-service, terraform-service, etc.) permet d&apos;avancer service par service et de maintenir une responsabilité claire par brique.
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
        </Box>
      )
    case 'ansible':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Module Ansible</Typography>
          <Typography>
            Le module Ansible se connecte à un serveur <strong>AWX</strong> ou <strong>Ansible Tower</strong>. Il permet de lister les job templates, de lancer des jobs et de consulter leur statut et leur historique.
          </Typography>
          <Typography component="h2">Configuration</Typography>
          <Typography>
            Configurez l&apos;URL du serveur AWX/Tower et les credentials (token ou identifiants) dans les paramètres du projet ou via les variables d&apos;environnement du service. Le frontend envoie les requêtes au backend (ansible-service ou proxy), qui communique avec l&apos;API AWX/Tower.
          </Typography>
          <Typography component="h2">Fonctionnalités</Typography>
          <Typography>
            Depuis l&apos;interface vous pouvez : parcourir les job templates disponibles, lancer un job sur un template, suivre l&apos;exécution (statut, logs si exposés), et consulter l&apos;historique des jobs. Les jobs sont typiquement associés au projet courant pour le filtrage et les droits.
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
          <Typography component="h2">Intégration</Typography>
          <Typography>
            Le pipeline-service peut recevoir des webhooks des plateformes CI/CD, enregistrer les exécutions et les afficher dans l&apos;interface. La corrélation avec les déploiements (K8s, Terraform) et les alertes est assurée via le bus d&apos;événements (Kafka) lorsque les services publient les événements correspondants.
          </Typography>
        </Box>
      )
    case 'users':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Utilisateurs et accès en production</Typography>
          <Typography>
            Ce qui suit décrit le comportement des utilisateurs une fois la solution mise en production et hébergée.
          </Typography>
          <Typography component="h2">Qui sont les utilisateurs ?</Typography>
          <Typography>
            Toute personne disposant d&apos;un <strong>compte</strong> peut utiliser la plateforme. Aujourd&apos;hui, l&apos;inscription est ouverte via la page Inscription (email, nom d&apos;utilisateur, mot de passe). En production vous pouvez : garder l&apos;inscription ouverte, ou la désactiver et créer les comptes via l&apos;API ou un outil admin. Le backend gère les rôles (admin, user) ; par défaut un nouvel inscrit a le rôle user.
          </Typography>
          <Typography component="h2">Accès aux données</Typography>
          <Typography>
            Chaque utilisateur a une liste de <strong>projets</strong> auxquels il a accès (projets qu&apos;il a créés ou auxquels il a été ajouté comme membre). Le projet courant est choisi sur la page Projets. Toutes les actions (Kubernetes, Terraform, Ansible, Pipelines) sont faites dans le cadre de ce projet. Les données sont isolées par <strong>project_id</strong> : un utilisateur ne voit que les projets dont il est membre.
          </Typography>
          <Typography component="h2">Travail en équipe</Typography>
          <Typography>
            Un projet a un <strong>propriétaire</strong> et peut avoir des <strong>membres</strong> (rôles admin ou member). L&apos;API expose l&apos;ajout de membre (<code>POST /api/v1/projects/:id/members</code>) et la liste des membres. Exposer une action « Inviter un membre » dans l&apos;interface permet aux équipes de gérer elles-mêmes les accès. Plusieurs personnes peuvent ainsi utiliser la même solution sur un même projet.
          </Typography>
          <Typography component="h2">Parcours type en production</Typography>
          <Box component="ul">
            <li>Accès à l&apos;URL hébergée (ex. https://kura.votredomaine.com)</li>
            <li>Compte : inscription (si ouverte) ou réception d&apos;un compte créé par un admin</li>
            <li>Connexion : email + mot de passe</li>
            <li>Projet : création d&apos;un nouveau projet ou sélection d&apos;un projet existant</li>
            <li>Modules : utilisation des modules dans le cadre du projet sélectionné</li>
          </Box>
          <Typography component="h2">À prévoir côté hébergement</Typography>
          <Typography>
            HTTPS obligatoire, sauvegarde régulière de PostgreSQL (utilisateurs, projets, membres), gestion sécurisée des secrets (JWT, clés cloud). Décider de la politique d&apos;inscription et prévoir un moyen de création de comptes si l&apos;inscription est réservée aux admins.
          </Typography>
        </Box>
      )
    case 'gke':
      return (
        <Box sx={contentSx}>
          <Typography component="h1">Connexion d&apos;un cluster GKE</Typography>
          <Typography>
            Pour connecter un cluster <strong>Google Kubernetes Engine (GKE)</strong> à Kura, vous pouvez soit fournir un kubeconfig (avec certificats complets), soit utiliser des credentials cloud (clé JSON du compte de service GCP) pour que le plugin GKE fonctionne côté serveur.
          </Typography>
          <Typography component="h2">1. Récupérer le kubeconfig</Typography>
          <Typography>
            Avec <code>gcloud</code> (Cloud Shell ou machine où gcloud est installé) :
          </Typography>
          <CodeBlock language="bash" label="Commande">
            {`gcloud container clusters get-credentials <NOM_CLUSTER> --region <REGION> --project <PROJECT_ID>`}
          </CodeBlock>
          <Typography>
            Copiez le contenu <strong>complet</strong> du fichier kubeconfig (avec certificats). Sous PowerShell : <code>Get-Content $env:USERPROFILE\.kube\config -Raw</code> ou <code>kubectl config view --raw</code>. Ne pas utiliser <code>kubectl config view</code> seul (sans <code>--raw</code>) : les certificats seraient masqués (DATA+OMITTED) et Kura ne pourrait pas se connecter.
          </Typography>
          <Typography component="h2">2. Dans Kura (sans Docker)</Typography>
          <Typography>
            Allez dans <strong>Kubernetes</strong> → <strong>Clusters</strong> → <strong>Ajouter un cluster</strong>. Choisissez le type GKE, collez le kubeconfig, et enregistrez. Le plugin GKE est inclus dans le k8s-service.
          </Typography>
          <Typography component="h2">3. En Docker : clé GCP pour GKE</Typography>
          <Typography>
            En Docker, le conteneur n&apos;a pas vos identifiants gcloud. Il faut un <strong>fichier de clé JSON d&apos;un compte de service GCP</strong> avec accès au cluster GKE (rôle conseillé : Utilisateur Kubernetes Engine). Créez la clé dans la console GCP (IAM → Comptes de service → Clés), placez le fichier par exemple dans <code>secrets/gcp-sa.json</code> (à ne pas commiter), montez-le dans le conteneur et définissez <code>GOOGLE_APPLICATION_CREDENTIALS=/secrets/gcp-sa.json</code>. Redémarrez au moins le service K8s après avoir monté la clé.
          </Typography>
          <Typography>
            Dans Kura : Kubernetes → Clusters → ajoutez le cluster en collant le kubeconfig (étape 1) puis activez le cluster. Si vous ne souhaitez pas utiliser de clé dans Docker, vous pouvez générer un kubeconfig avec token (voir la doc GKE « Cluster access for kubectl ») ; le token devra être renouvelé régulièrement.
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
  const [selectedId, setSelectedId] = useState('intro')
  const [sidebarOpen, setSidebarOpen] = useState(false)
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
                <ListItemButton
                  selected={selectedId === section.id}
                  onClick={() => {
                    setSelectedId(section.id)
                    if (isMobile) setSidebarOpen(false)
                  }}
                  sx={{
                    py: 0.75,
                    '&.Mui-selected': { bgcolor: 'rgba(0,229,255,0.15)', borderRight: `3px solid ${jellyfishColors.cyanSoft}` },
                  }}
                >
                  <ListItemText primary={section.label} primaryTypographyProps={{ fontSize: '0.9rem', fontWeight: 600 }} />
                </ListItemButton>
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
                      '&.Mui-selected': { bgcolor: 'rgba(0,229,255,0.12)', borderRight: `3px solid ${jellyfishColors.cyanSoft}` },
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
                  '&.Mui-selected': { bgcolor: 'rgba(0,229,255,0.15)', borderRight: `3px solid ${jellyfishColors.cyanSoft}` },
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
            boxShadow: `inset 0 1px 0 rgba(0,229,255,0.08)`,
          }}
        >
          <DocContent docId={selectedId} />
        </Paper>
      </Box>
    </Box>
  )
}
