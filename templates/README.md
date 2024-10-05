This template can be served to users instead of the React app in this repo.

To use it instead of the React app, copy this function to handlers.go:

```go
// serveTemplate renders the index.html template
func serveTemplate(w http.ResponseWriter, _ *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
```

Then in the `main.go` file, add the following route:

```go
http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
 	serveTemplate(w, r)
})
```

This will serve the template at the root of your application.

