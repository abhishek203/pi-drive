package api

import (
	"net/http"
	"time"

	"github.com/pidrive/pidrive/internal/activity"
	"github.com/pidrive/pidrive/internal/auth"
	"github.com/pidrive/pidrive/internal/billing"
	"github.com/pidrive/pidrive/internal/config"
	"github.com/pidrive/pidrive/internal/db"
	"github.com/pidrive/pidrive/internal/mount"
	"github.com/pidrive/pidrive/internal/search"
	"github.com/pidrive/pidrive/internal/share"
	"github.com/pidrive/pidrive/internal/trash"
)

type Server struct {
	cfg             *config.Config
	db              *db.DB
	authService     *auth.AuthService
	emailService    *auth.EmailService
	shareService    *share.ShareService
	fileManager     *share.FileManager
	activityService *activity.ActivityService
	searchService   *search.SearchService
	indexer         *search.Indexer
	trashService    *trash.TrashService
	billingService  *billing.BillingService
	mountService    *mount.MountService
}

func NewServer(cfg *config.Config, database *db.DB) *Server {
	authSvc := auth.NewAuthService(database)
	emailSvc := auth.NewEmailService(cfg.ResendAPIKey, cfg.FromEmail)
	shareSvc := share.NewShareService(database, cfg.ServerURL)
	fileMgr := share.NewFileManager(cfg.JuiceFSMountPath)
	actSvc := activity.NewActivityService(database)
	searchSvc := search.NewSearchService(database)
	idx := search.NewIndexer(database, cfg.JuiceFSMountPath, 60*time.Second)
	trashSvc := trash.NewTrashService(database, cfg.JuiceFSMountPath, cfg.JuiceFSTrashDays)
	billingSvc := billing.NewBillingService(database)
	mountSvc := mount.NewMountService(cfg)

	return &Server{
		cfg:             cfg,
		db:              database,
		authService:     authSvc,
		emailService:    emailSvc,
		shareService:    shareSvc,
		fileManager:     fileMgr,
		activityService: actSvc,
		searchService:   searchSvc,
		indexer:         idx,
		trashService:    trashSvc,
		billingService:  billingSvc,
		mountService:    mountSvc,
	}
}

func (s *Server) Router() http.Handler {
	mux := http.NewServeMux()

	authHandler := NewAuthHandler(s.authService, s.emailService)
	mountHandler := NewMountHandler(s.mountService, s.fileManager, s.activityService)
	shareHandler := NewShareHandler(s.shareService, s.fileManager, s.authService, s.activityService, s.billingService)
	searchHandler := NewSearchHandler(s.searchService, s.indexer)
	activityHandler := NewActivityHandler(s.activityService)
	trashHandler := NewTrashHandler(s.trashService, s.activityService)
	billingHandler := NewBillingHandler(s.billingService)

	authMiddleware := AuthMiddleware(s.authService)

	// Public routes
	mux.HandleFunc("POST /api/register", authHandler.Register)
	mux.HandleFunc("POST /api/login", authHandler.Login)
	mux.HandleFunc("POST /api/verify", authHandler.Verify)
	mux.HandleFunc("GET /api/plans", billingHandler.GetPlans)
	mux.HandleFunc("GET /s/{id}", shareHandler.ServeShareLink)

	// Landing page & docs
	mux.HandleFunc("GET /{$}", s.serveLanding)
	mux.HandleFunc("GET /install.sh", s.serveInstallScript)
	mux.HandleFunc("GET /skill.md", s.serveSkillMD)
	mux.HandleFunc("GET /llms.txt", s.serveSkillMD)
	mux.HandleFunc("GET /robots.txt", s.serveRobots)
	mux.HandleFunc("GET /sitemap.xml", s.serveSitemap)
	mux.HandleFunc("GET /docs", s.serveDocs)
	mux.HandleFunc("GET /blog/", s.serveBlog)
	mux.HandleFunc("GET /vs/", s.serveVs)

	// WebDAV (all methods)
	mux.HandleFunc("/webdav/", s.serveWebDAV)
	mux.HandleFunc("/webdav", s.serveWebDAV)

	// Binary releases
	mux.HandleFunc("GET /releases/", s.serveRelease)

	// Protected routes
	protected := http.NewServeMux()
	protected.HandleFunc("GET /api/me", authHandler.Me)
	protected.HandleFunc("GET /api/whoami", authHandler.Me)
	protected.HandleFunc("POST /api/mount", mountHandler.Mount)
	protected.HandleFunc("POST /api/unmount", mountHandler.Unmount)
	protected.HandleFunc("POST /api/share", shareHandler.Share)
	protected.HandleFunc("POST /api/share/link", shareHandler.ShareLink)
	protected.HandleFunc("GET /api/shared", shareHandler.ListShared)
	protected.HandleFunc("DELETE /api/share/{id}", shareHandler.Revoke)
	protected.HandleFunc("GET /api/search", searchHandler.Search)
	protected.HandleFunc("POST /api/index", searchHandler.TriggerIndex)
	protected.HandleFunc("GET /api/activity", activityHandler.List)
	protected.HandleFunc("GET /api/trash", trashHandler.List)
	protected.HandleFunc("POST /api/trash/restore", trashHandler.Restore)
	protected.HandleFunc("DELETE /api/trash", trashHandler.Empty)
	protected.HandleFunc("GET /api/usage", billingHandler.GetUsage)
	protected.HandleFunc("POST /api/upgrade", billingHandler.Upgrade)
	protected.HandleFunc("GET /api/billing", billingHandler.GetBilling)

	mux.Handle("/api/", authMiddleware(protected))

	return mux
}

func (s *Server) StartIndexer() {
	s.indexer.Start()
}

func (s *Server) StopIndexer() {
	s.indexer.Stop()
}
