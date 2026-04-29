package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Addr              string
	DataDir           string
	DBPath            string
	UploadDir         string
	AdminUser         string
	AdminPass         string
	SessionTTL        time.Duration
	AllowedCIDRs      []string
	DeniedCIDRs       []string
	TrustProxyHeaders bool
	MaxMindDBPath     string
	AllowedCountries  []string
	DeniedCountries   []string
	MaxUploadBytes    int64
	PublicSiteName    string
}

func Load() Config {
	data := env("CMS_DATA_DIR", "./data")
	return Config{
		Addr:              env("CMS_ADDR", ":8080"),
		DataDir:           data,
		DBPath:            env("CMS_DB", data+"/cms.db"),
		UploadDir:         env("CMS_UPLOAD_DIR", data+"/uploads"),
		AdminUser:         env("CMS_ADMIN_USER", "admin"),
		AdminPass:         env("CMS_ADMIN_PASS", "change-me"),
		SessionTTL:        dur("CMS_SESSION_TTL", 12*time.Hour),
		AllowedCIDRs:      csv("CMS_ALLOW_CIDRS"),
		DeniedCIDRs:       csv("CMS_DENY_CIDRS"),
		TrustProxyHeaders: boolEnv("CMS_TRUST_PROXY_HEADERS", false),
		MaxMindDBPath:     os.Getenv("CMS_MAXMIND_DB"),
		AllowedCountries:  upperCSV("CMS_ALLOW_COUNTRIES"),
		DeniedCountries:   upperCSV("CMS_DENY_COUNTRIES"),
		MaxUploadBytes:    int64Env("CMS_MAX_UPLOAD_BYTES", 25<<20),
		PublicSiteName:    env("CMS_SITE_NAME", "TinyCMS"),
	}
}

func env(k, d string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return d
}
func csv(k string) []string      { return split(os.Getenv(k), false) }
func upperCSV(k string) []string { return split(strings.ToUpper(os.Getenv(k)), true) }
func split(s string, upper bool) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			if upper {
				p = strings.ToUpper(p)
			}
			out = append(out, p)
		}
	}
	return out
}
func boolEnv(k string, d bool) bool {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return d
	}
	return b
}
func int64Env(k string, d int64) int64 {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return d
	}
	return n
}
func dur(k string, d time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return d
	}
	x, err := time.ParseDuration(v)
	if err != nil {
		return d
	}
	return x
}
