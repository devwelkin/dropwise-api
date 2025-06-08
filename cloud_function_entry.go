// ./cloud_function_entry.go (for deploy)
package functionrunner

import (
	"net/http"

	"github.com/nouvadev/dropwise/internal/worker"
)

func ActualEntryPoint(w http.ResponseWriter, r *http.Request) {
	worker.ProcessDueDropsHTTP(w, r)
}
