package leads

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler serves the API and compiled frontend without logging request data or PII.
func Handler(store *Store, assets fs.FS) http.Handler {
	router := gin.New()
	router.Use(safeRecovery)
	router.GET("/api/leads", store.listHTTP)
	router.POST("/api/leads", store.addHTTP)
	router.POST("/api/leads/import", store.importHTTP)
	files := http.FileServer(http.FS(assets))
	router.NoRoute(func(ctx *gin.Context) {
		if strings.HasPrefix(ctx.Request.URL.Path, "/api/") {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "endpoint not found"})
			return
		}
		ctx.Header("X-Content-Type-Options", "nosniff")
		files.ServeHTTP(ctx.Writer, ctx.Request)
	})
	return router
}

func safeRecovery(ctx *gin.Context) {
	defer func() {
		if recover() != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
	}()
	ctx.Header("Cache-Control", "no-store")
	ctx.Next()
}

func (store *Store) listHTTP(ctx *gin.Context) {
	leads, err := store.List(ctx.Request.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not load leads"})
		return
	}
	ctx.JSON(http.StatusOK, leads)
}

func (store *Store) addHTTP(ctx *gin.Context) {
	const maxJSONBytes = 64 << 10
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var lead *Lead
	if err := decoder.Decode(&lead); err != nil || lead == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "body must be a valid lead JSON object"})
		return
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "body must contain a single JSON object"})
		return
	}
	lead.Normalize()
	if fields := lead.Validate(time.Now()); len(fields) > 0 {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"errors": fields})
		return
	}
	saved, err := store.Add(ctx.Request.Context(), *lead)
	if errors.Is(err, ErrDuplicateEmail) {
		ctx.JSON(http.StatusConflict, gin.H{"errors": gin.H{"email": "already exists"}})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not save lead"})
		return
	}
	ctx.JSON(http.StatusCreated, saved)
}

func (store *Store) importHTTP(ctx *gin.Context) {
	const maxCSVBytes = 10 << 20
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxCSVBytes)
	result, err := store.Import(ctx.Request.Context(), ctx.Request.Body)
	if err != nil {
		// Include progress because earlier valid rows remain committed.
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "import stopped; check headers, CSV format, and the 10 MB limit", "result": result})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
