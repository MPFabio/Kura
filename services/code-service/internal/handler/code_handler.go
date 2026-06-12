package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/modulops/code-service/internal/service"
)

// CodeHandler expose les endpoints HTTP du module Code.
type CodeHandler struct {
	svc *service.CodeService
}

// New crée un CodeHandler.
func New(svc *service.CodeService) *CodeHandler {
	return &CodeHandler{svc: svc}
}

// ListRepositories renvoie les dépôts GitHub liés à un projet.
func (h *CodeHandler) ListRepositories(c *gin.Context) {
	projectID := c.Query("project_id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramètre project_id requis"})
		return
	}

	repos, err := h.svc.ListRepositories(c.Request.Context(), c.GetHeader("Authorization"), projectID)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": repos})
}

// GetTree renvoie le contenu d'un répertoire d'un dépôt.
func (h *CodeHandler) GetTree(c *gin.Context) {
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramètre repo requis"})
		return
	}

	entries, err := h.svc.GetTree(c.Request.Context(), repo, c.Query("path"), c.Query("ref"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GetFile renvoie le contenu d'un fichier d'un dépôt.
func (h *CodeHandler) GetFile(c *gin.Context) {
	repo := c.Query("repo")
	path := c.Query("path")
	if repo == "" || path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramètres repo et path requis"})
		return
	}

	file, err := h.svc.GetFile(c.Request.Context(), repo, path, c.Query("ref"))
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, file)
}

// GetCommits renvoie l'historique des commits d'un dépôt.
func (h *CodeHandler) GetCommits(c *gin.Context) {
	repo := c.Query("repo")
	if repo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramètre repo requis"})
		return
	}

	page := 0
	if p := c.Query("page"); p != "" {
		if v, err := parsePositiveInt(p); err == nil {
			page = v
		}
	}

	commits, err := h.svc.GetCommits(c.Request.Context(), repo, c.Query("path"), c.Query("ref"), page)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, commits)
}

// GetCommitDiff renvoie le détail d'un commit (fichiers modifiés + diffs).
func (h *CodeHandler) GetCommitDiff(c *gin.Context) {
	repo := c.Query("repo")
	sha := c.Param("sha")
	if repo == "" || sha == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "paramètres repo et sha requis"})
		return
	}

	detail, err := h.svc.GetCommitDiff(c.Request.Context(), repo, sha)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, detail)
}

func parsePositiveInt(s string) (int, error) {
	var v int
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, &invalidIntError{s}
		}
		v = v*10 + int(r-'0')
	}
	return v, nil
}

type invalidIntError struct{ s string }

func (e *invalidIntError) Error() string { return "valeur entière invalide: " + e.s }
