package context

import (
	"context"
	"net/http"
)

type Handler struct {
	BundleRootDir string
	Handler http.Handler
}

const KEY_BUNDLE_ROOT_DIR = "BUNDLE_ROOT_DIR"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	old_context := r.Context()
	new_context := context.WithValue(
		old_context, 
		KEY_BUNDLE_ROOT_DIR, 
		h.BundleRootDir)
	r_new := r.WithContext(new_context)
	h.Handler.ServeHTTP(w , r_new)
}

// BundleDir extracts the bundle dir from ctx, if present.
func BundleRootDir(ctx context.Context) string {
	// ctx.Value returns nil if ctx has no value for the key;
	// the string type assertion returns ok=false for nil.
	bundle_dir, _ := ctx.Value(KEY_BUNDLE_ROOT_DIR).(string)
	return bundle_dir	
}


