package geo

import (
	"net/http"
	"strings"

	"github.com/oschwald/geoip2-golang"
	"tinycms/internal/auth"
)

type Filter struct {
	DB          *geoip2.Reader
	Allow, Deny map[string]bool
	TrustProxy  bool
}

func New(dbPath string, allow, deny []string, trustProxy bool) (*Filter, error) {
	if dbPath == "" {
		return &Filter{TrustProxy: trustProxy}, nil
	}
	db, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &Filter{DB: db, Allow: set(allow), Deny: set(deny), TrustProxy: trustProxy}, nil
}
func set(xs []string) map[string]bool {
	m := map[string]bool{}
	for _, x := range xs {
		m[strings.ToUpper(strings.TrimSpace(x))] = true
	}
	return m
}

func (f *Filter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if f == nil || f.DB == nil {
			next.ServeHTTP(w, r)
			return
		}
		ip := auth.ClientIP(r, f.TrustProxy)
		rec, err := f.DB.Country(ip)
		if err != nil {
			http.Error(w, "geo denied", http.StatusForbidden)
			return
		}
		cc := strings.ToUpper(rec.Country.IsoCode)
		if f.Deny[cc] || (len(f.Allow) > 0 && !f.Allow[cc]) {
			http.Error(w, "geo denied", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (f *Filter) Close() error {
	if f != nil && f.DB != nil {
		return f.DB.Close()
	}
	return nil
}
