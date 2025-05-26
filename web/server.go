package web

import (
	"log"
	"log/slog"
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func Navbar() Node {
	return Nav(Class("navbar"),
		Ol(
			Li(A(Href("/"), Text("Home"))),
			Li(A(Href("/contact"), Text("Contact"))),
			Li(A(Href("/about"), Text("About"))),
		),
	)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	Navbar().Render(w)
}

func Start() {
	// Create a new ServeMux to use our middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)

	// Wrap the mux with the pretty print middleware
	handler := PrettyPrintHTML(mux)

	port := ":8080"
	slog.Info("Starting server on", "port", port)

	log.Fatal(http.ListenAndServe(port, handler))
}
