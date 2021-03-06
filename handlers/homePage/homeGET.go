package homePage

import (
	"net/http"
	"time"

	"go-triton-app/app"
)

var indexView = app.TemplateManager.MustParseLocalizedView("home.html")

func HomeGET(w http.ResponseWriter, r *http.Request) {
	resp := app.HTMLResponse(w, r)

	pageData := &HomePageData{Time: time.Now().String()}
	pageHTML := indexView.MustExecuteToString(resp.Lang(), pageData)

	d := app.MasterPageData(resp.NewLocalizedTitle("home"), pageHTML)
	resp.MustComplete(d)
}
