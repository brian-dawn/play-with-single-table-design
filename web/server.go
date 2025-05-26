package web

import (
	"log"
	"log/slog"
	"net/http"

	. "maragu.dev/gomponents"
	. "maragu.dev/gomponents/html"
)

func BaseHTML(content Node) Node {
	return HTML(
		Lang("en"),
		Head(
			Meta(Charset("utf-8")),
			Meta(Name("viewport"), Content("width=device-width, initial-scale=1.0")),
			Title("Your App"),
			// Tailwind CSS CDN
			Script(Src("https://cdn.tailwindcss.com")),
			// HTMX CDN
			Script(Src("https://unpkg.com/htmx.org@1.9.10")),
			// Add HTMX attributes to body
			Script(Raw(`
				htmx.config = {
					defaultSwapStyle: 'innerHTML'
				}
			`)),
		),
		Body(
			Class("min-h-screen bg-gray-50"),
			Div(
				Class("mx-auto max-w-3xl px-4 sm:px-6 lg:px-8"), // Container with responsive padding
				content,
			),
		),
	)
}

func Navbar() Node {
	return Nav(
		Class("sticky top-0 bg-white shadow-sm mb-8"),
		Div(
			Class("mx-auto max-w-3xl px-4 sm:px-6 lg:px-8"), // Match container width
			Div(
				Class("flex h-16 items-center justify-between"),
				// Logo/Brand
				A(
					Href("/"),
					Class("text-xl font-semibold text-gray-900"),
					Text("Your App"),
				),
				// Navigation items
				Div(
					Class("hidden sm:block"), // Hide on mobile
					Ol(
						Class("flex space-x-8"),
						Li(A(Href("/"), Class("text-gray-700 hover:text-blue-600 transition-colors"), Text("Home"))),
						Li(A(Href("/contact"), Class("text-gray-700 hover:text-blue-600 transition-colors"), Text("Contact"))),
						Li(A(Href("/about"), Class("text-gray-700 hover:text-blue-600 transition-colors"), Text("About"))),
					),
				),
				// Mobile menu button
				Button(
					Type("button"),
					Class("sm:hidden p-2 text-gray-700 hover:text-blue-600"),
					Attr("aria-label", "Toggle menu"),
					Text("☰"),
				),
			),
		),
		// Mobile menu (hidden by default)
		Div(
			Class("sm:hidden hidden"), // Initially hidden, toggle with HTMX
			Attr("id", "mobile-menu"),
			Ol(
				Class("flex flex-col space-y-4 px-4 py-6"),
				Li(A(Href("/"), Class("text-gray-700 hover:text-blue-600 block transition-colors"), Text("Home"))),
				Li(A(Href("/contact"), Class("text-gray-700 hover:text-blue-600 block transition-colors"), Text("Contact"))),
				Li(A(Href("/about"), Class("text-gray-700 hover:text-blue-600 block transition-colors"), Text("About"))),
			),
		),
	)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte("<!DOCTYPE html>\n"))
	BaseHTML(Navbar()).Render(w)
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
