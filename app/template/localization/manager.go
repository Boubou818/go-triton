package localization

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"golang.org/x/text/language"

	"github.com/mgenware/go-packagex/filepathx"
	"github.com/mgenware/go-triton/app/defs"
)

var (
	LanguageCSTag = language.SimplifiedChinese
	LanguageENTag = language.English
)

var matcher = language.NewMatcher([]language.Tag{
	LanguageENTag, // The first language is used as fallback.
	LanguageCSTag,
})

type Manager struct {
	defaultDic *Dictionary
	dics       map[string]*Dictionary
}

// NewManagerFromDirectory creates a Manager from a directory of translation files.
func NewManagerFromDirectory(dir string, defaultLang string) (*Manager, error) {
	fileNames, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	dics := make(map[string]*Dictionary)
	for _, info := range fileNames {
		if !info.IsDir() {
			d, err := NewDictionaryFromFile(filepath.Join(dir, info.Name()))
			if err != nil {
				return nil, err
			}

			name := filepathx.TrimExt(info.Name())
			dics[name] = d
			log.Printf("Read localization file \"%v\"", name)
		}
	}
	if len(dics) == 0 {
		return nil, fmt.Errorf("No dictionary found in %v", dir)
	}

	defaultDic := dics[defaultLang]
	if defaultDic == nil {
		return nil, fmt.Errorf("Default language \"%v\" not found", defaultLang)
	}

	return &Manager{dics: dics, defaultDic: defaultDic}, nil
}

// DictionaryForLanguage returns an Dictionary object associated with the specified language.
func (mgr *Manager) DictionaryForLanguage(lang string) *Dictionary {
	dic := mgr.dics[lang]
	if dic == nil {
		return mgr.defaultDic
	}
	return dic
}

// ValueForKeyWithLanguage returns a localized string associated with the specified language and key.
func (mgr *Manager) ValueForKeyWithLanguage(lang, key string) string {
	dic := mgr.DictionaryForLanguage(lang)
	if dic == nil {
		return ""
	}
	return dic.Map[key]
}

// ValueForKey returns a localized string associated with the specified context and key.
func (mgr *Manager) ValueForKey(ctx context.Context, key string) string {
	return mgr.ValueForKeyWithLanguage(defs.ContextLanguage(ctx), key)
}

// MatchLanguage returns the determined language based on various conditions.
func (mgr *Manager) MatchLanguage(ctx context.Context, w http.ResponseWriter, r *http.Request) string {
	// Check if user has explicitly set a language
	queryLang := r.FormValue(defs.LanguageQueryKey)
	if queryLang != "" {
		// Write the user specified language to cookies
		expires := time.Now().Add(30 * 24 * time.Hour)
		c := &http.Cookie{Name: defs.LanguageCookieKey, Value: queryLang, Expires: expires}
		http.SetCookie(w, c)

		return queryLang
	}

	// If no user-specified language exists, try to use the cookie value
	cookieLang, _ := r.Cookie(defs.LanguageCookieKey)
	if cookieLang != nil {
		return cookieLang.Value
	}

	// If none of the above values exist, use the language matcher
	accept := r.Header.Get("Accept-Language")
	_, index := language.MatchStrings(matcher, accept)

	if index == 1 {
		return defs.LanguageCSString
	}

	// Fallback to English
	return defs.LanguageENString
}

// EnableContextLanguage defines a middleware to set the context language associated with the request.
func (mgr *Manager) EnableContextLanguage(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		lang := mgr.MatchLanguage(ctx, w, r)
		ctx = context.WithValue(ctx, defs.LanguageContextKey, lang)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
