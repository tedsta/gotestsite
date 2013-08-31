package gotestsite

import (
	"fmt"
	"net/http"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h1>Hello, world!</h1>")
}

func main() {
	http.HandleFunc("/", handler)
	http.ListenAndServe("", nil)
}
