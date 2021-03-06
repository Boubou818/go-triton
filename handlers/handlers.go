package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mgenware/go-packagex/iox"

	"go-triton-app/app"
	"go-triton-app/handlers/errorPage"
	"go-triton-app/handlers/homePage"
	"go-triton-app/handlers/system"

	"github.com/go-chi/chi"
)

// Start starts the web router.
func Start() {
	r := chi.NewRouter()
	config := app.Config
	httpConfig := config.HTTP

	// ----------------- Middlewares -----------------
	// THE PanicMiddleware MUST BE AT THE VERY BEGINNING, OTHERWISE IT WILL NOT WORK!
	if !config.DevMode {
		// *** Production only ***

		// Mount PanicMiddleware only in production, let panic crash in development
		r.Use(system.PanicMiddleware)
	}

	// Mount static file server
	httpStaticConfig := httpConfig.Static
	if httpStaticConfig != nil {
		pattern := httpStaticConfig.Pattern
		dir := httpStaticConfig.Dir
		log.Printf("✅ Serving Assets(%v) at \"%v\"", pattern, dir)
		fileServer(r, pattern, http.Dir(dir))
		if !iox.IsDirectory(dir) {
			log.Printf("☢️ Assets directory \"%v\" doesn't exist", dir)
		}
	}

	// Mount other middlewares, for example:
	// r.Use(sessionMiddleware)

	// ----------------- HTTP Routes -----------------
	lm := app.TemplateManager.LocalizationManager

	// Not found handler
	r.With(lm.EnableContextLanguage).NotFound(system.NotFoundHandler)

	// index handler
	r.With(lm.EnableContextLanguage).Get("/", homePage.HomeGET)
	r.With(lm.EnableContextLanguage).Get("/fakeError", errorPage.FakeErrorGET)

	log.Printf("🚙 Server running at %v", httpConfig.Port)
	err := http.ListenAndServe(":"+strconv.Itoa(httpConfig.Port), r)
	if err != nil {
		panic(err)
	}
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	fs := http.StripPrefix(path, http.FileServer(root))

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	}))
}
