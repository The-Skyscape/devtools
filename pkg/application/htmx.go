package application

import (
	"net/http"
	"strings"
)

// IsHTMX checks if the request is from HTMX
func IsHTMX(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// HTMXRedirect sends an HTMX redirect response
func HTMXRedirect(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusNoContent)
}

// HTMXRefresh triggers a page refresh for HTMX
func HTMXRefresh(w http.ResponseWriter) {
	w.Header().Set("HX-Refresh", "true")
	w.WriteHeader(http.StatusNoContent)
}

// HTMXReswap changes the swap behavior for the current request
func HTMXReswap(w http.ResponseWriter, swapType string) {
	w.Header().Set("HX-Reswap", swapType)
}

// HTMXRetarget changes the target element for the current request
func HTMXRetarget(w http.ResponseWriter, selector string) {
	w.Header().Set("HX-Retarget", selector)
}

// HTMXTrigger triggers client-side events
func HTMXTrigger(w http.ResponseWriter, events ...string) {
	w.Header().Set("HX-Trigger", strings.Join(events, ","))
}

// HTMXPushURL pushes a new URL to the browser history
func HTMXPushURL(w http.ResponseWriter, url string) {
	w.Header().Set("HX-Push-URL", url)
}

// CurrentURL gets the current URL from HTMX headers
func CurrentURL(r *http.Request) string {
	if url := r.Header.Get("HX-Current-URL"); url != "" {
		return url
	}
	return r.URL.String()
}

// IsBootsed checks if the request is boosted by HTMX
func IsBoosted(r *http.Request) bool {
	return r.Header.Get("HX-Boosted") == "true"
}

// HistoryRestoreRequest checks if this is a history restore request
func IsHistoryRestore(r *http.Request) bool {
	return r.Header.Get("HX-History-Restore-Request") == "true"
}