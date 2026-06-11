package app

import (
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"server-management-service/internal/infrastructure/gateway"
	"server-management-service/internal/infrastructure/security"
	resthandler "server-management-service/internal/modules/server_management/handler/rest"
)

// setupHTTPRoutes builds the HTTP mux with all routes.
func (a *App) setupHTTPRoutes(gwmux *runtime.ServeMux) *http.ServeMux {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// OpenAPI JSON specs
	mux.Handle("/openapi/", http.StripPrefix("/openapi/", http.FileServer(http.Dir("./api/openapi"))))

	// Swagger UI
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./docs/swagger-ui.html")
	})

	// REST import/export
	if a.RESTImportExport != nil {
		importHandler := http.HandlerFunc(a.RESTImportExport.HandleImport)
		exportHandler := http.HandlerFunc(a.RESTImportExport.HandleExport)

		mux.Handle("/api/v1/servers/import",
			resthandler.AuthMiddleware(a.Authenticator)(
				resthandler.PermissionMiddleware(a.Authorizer, security.PermServerImport)(
					importHandler,
				),
			),
		)
		mux.Handle("/api/v1/servers/export",
			resthandler.AuthMiddleware(a.Authenticator)(
				resthandler.PermissionMiddleware(a.Authorizer, security.PermServerExport)(
					exportHandler,
				),
			),
		)
	}

	// gRPC-gateway catch-all (with cookie middleware)
	mux.Handle("/", gateway.CookieMiddleware(gwmux))

	return mux
}
