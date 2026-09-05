package api

import (
	"net"
	"net/http"

	"web-app-firewall-ml-detection/internal/config"
	"web-app-firewall-ml-detection/internal/middleware"
	"web-app-firewall-ml-detection/internal/proxy"
)

func NewRouter(
	cfg *config.Config,
	wafHandler *proxy.WAFHandler,
	authHandler *AuthHandler,
	domainHandler *DomainHandler,
	ruleHandler *RuleHandler,
	dnsHandler *DNSHandler,
	logHandler *LogHandler,
	systemHandler *SystemHandler,
) http.Handler {

	apiMux := http.NewServeMux()

	// --- Auth Routes ---
	apiMux.HandleFunc("/api/auth/register", authHandler.Register)
	apiMux.HandleFunc("/api/auth/login", authHandler.Login)
	apiMux.HandleFunc("/api/auth/logout", authHandler.Logout)
	apiMux.HandleFunc("/api/system/status", systemHandler.GetSystemStatus)
	apiMux.HandleFunc("/api/system/traffic-history", authHandler.Middleware(systemHandler.GetTrafficHistory))

	apiMux.HandleFunc("/api/auth/check", authHandler.Middleware(authHandler.CheckAuth))
	apiMux.HandleFunc("/api/auth/verify", authHandler.VerifyEmail)
	apiMux.HandleFunc("/api/auth/email/update", authHandler.Middleware(authHandler.RequestEmailChange))
	apiMux.HandleFunc("/api/auth/email/verify-change", authHandler.VerifyEmailChange)
	apiMux.HandleFunc("/api/auth/password/update", authHandler.Middleware(authHandler.UpdatePassword))
	apiMux.HandleFunc("/api/auth/password/forgot", authHandler.ForgotPassword)
	apiMux.HandleFunc("/api/auth/password/reset", authHandler.ResetPassword)

	// --- Domain Routes ---
	apiMux.HandleFunc("/api/domains", authHandler.Middleware(domainHandler.ListDomains))
	apiMux.HandleFunc("/api/domains/add", authHandler.Middleware(domainHandler.AddDomain))
	apiMux.HandleFunc("/api/domains/verify", authHandler.Middleware(domainHandler.Verify))
	apiMux.HandleFunc("/api/domains/delete", authHandler.Middleware(domainHandler.DeleteDomain))

	// --- DNS Routes ---
	apiMux.HandleFunc("/api/dns/records", authHandler.Middleware(dnsHandler.ManageRecords))

	// --- Rule Routes ---
	apiMux.HandleFunc("/api/rules/global", authHandler.Middleware(ruleHandler.GetGlobal))
	apiMux.HandleFunc("/api/rules/custom", authHandler.Middleware(ruleHandler.GetCustom))
	apiMux.HandleFunc("/api/rules/custom/add", authHandler.Middleware(ruleHandler.AddCustom))
	apiMux.HandleFunc("/api/rules/toggle", authHandler.Middleware(ruleHandler.Toggle))

	// --- Log Routes ---
	apiMux.HandleFunc("/api/logs", authHandler.Middleware(logHandler.GetLogs))
	apiMux.HandleFunc("/api/logs/stream", logHandler.SSEHandler)

	// ------- HOST-BASED DISPATCH -------
	root := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if h, _, err := net.SplitHostPort(host); err == nil {
			host = h
		}

		switch host {
		case "api.minishield.tech", "minishield.tech", "www.minishield.tech":
			apiMux.ServeHTTP(w, r) 
		default:
			wafHandler.ServeHTTP(w, r) // customer traffic → WAF proxy
		}
	})

	return middleware.CORS(cfg)(root)
}