// ./cloud_function_entry.go (go.mod'un yanında)
package functionrunner // Google Cloud Functions genellikle 'main' paketini sever

import (
	"net/http"
	// worker paketinin yeni konumu için doğru import yolu:
	"github.com/twomotive/dropwise/worker"
)

// Cloud Function tarafından çağrılacak olan ASIL fonksiyon bu olacak.
// gcloud komutundaki --entry-point buna göre güncellenmeli.
func ActualEntryPoint(w http.ResponseWriter, r *http.Request) {
	worker.ProcessDueDropsHTTP(w, r) // Asıl worker fonksiyonunu çağırır
}
