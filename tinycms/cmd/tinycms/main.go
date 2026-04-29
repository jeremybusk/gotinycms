package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"tinycms/cmsv1connect"
	"tinycms/internal/auth"
	"tinycms/internal/config"
	"tinycms/internal/db"
	"tinycms/internal/geo"
	"tinycms/internal/service"
	"tinycms/internal/web"
)

func main() {
	cfg := config.Load()
	must(os.MkdirAll(cfg.UploadDir, 0750))
	store, err := db.Open(cfg.DBPath)
	must(err)
	ipf, err := auth.NewIPFilter(cfg.AllowedCIDRs, cfg.DeniedCIDRs, cfg.TrustProxyHeaders)
	must(err)
	geof, err := geo.New(cfg.MaxMindDBPath, cfg.AllowedCountries, cfg.DeniedCountries, cfg.TrustProxyHeaders)
	must(err)
	defer geof.Close()

	svc := &service.Service{Store: store, UploadDir: cfg.UploadDir, MaxUploadBytes: cfg.MaxUploadBytes, SiteName: cfg.PublicSiteName}
	_, api := cmsv1connect.NewCMSServiceHandler(svc)
	admin := http.FileServer(http.Dir("web/dist"))
	uploads := http.StripPrefix("/uploads/", http.FileServer(http.Dir(cfg.UploadDir)))
	pub := web.NewPublic(store, cfg.PublicSiteName)

	mux := http.NewServeMux()
	mux.Handle("/cms.v1.CMSService/", chain(api, ipf.Middleware, geof.Middleware, auth.Basic{User: cfg.AdminUser, Pass: cfg.AdminPass}.Middleware))
	mux.Handle("/uploads/", chain(uploads, ipf.Middleware, geof.Middleware))
	mux.Handle("/admin/", chain(http.StripPrefix("/admin/", admin), ipf.Middleware, geof.Middleware, auth.Basic{User: cfg.AdminUser, Pass: cfg.AdminPass}.Middleware))
	mux.Handle("/", chain(pub, ipf.Middleware, geof.Middleware))

	srv := &http.Server{Addr: cfg.Addr, Handler: secureHeaders(mux), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 30 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 120 * time.Second, MaxHeaderBytes: 1 << 20}
	log.Printf("tinycms listening on %s db=%s uploads=%s", cfg.Addr, cfg.DBPath, filepath.Clean(cfg.UploadDir))
	log.Fatal(srv.ListenAndServe())
}

type mw func(http.Handler) http.Handler

func chain(h http.Handler, mws ...mw) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}
func must(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
