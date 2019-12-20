package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/auth0-community/auth0"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	jose "gopkg.in/square/go-jose.v2"
)

// Product contains information about VR experiences
type Product struct {
	ID          int
	Name        string
	Slug        string
	Description string
}

// Catalog of VR experiences and store them in a slice
var products = []Product{
	Product{ID: 1, Name: "Hover Shooters", Slug: "hover-shooters", Description: "Shoot your way to the top on 14 different hoverboards"},
	Product{ID: 2, Name: "Ocean Explorer", Slug: "ocean-explorer", Description: "Explore the depths of the sea in this one of a kind underwater experience"},
	Product{ID: 3, Name: "Dinosaur Park", Slug: "dinosaur-park", Description: "Go back 65 million years in the past and ride a T-Rex"},
	Product{ID: 4, Name: "Cars VR", Slug: "cars-vr", Description: "Get behind the wheel of the fastest cars in the world."},
	Product{ID: 5, Name: "Robin Hood", Slug: "robin-hood", Description: "Pick up the bow and arrow and master the art of archery"},
	Product{ID: 6, Name: "Real World VR", Slug: "real-world-vr", Description: "Explore the seven wonders of the world in VR"},
}

// ProductsHandler called when user makes GET request to /products route. Returns a list of products available for review
var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Convert slice of products to JSON
	payload, _ := json.Marshal(products)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
})

// AddFeedbackHandler will add either positive or negative feedback to the product
var AddFeedbackHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var product Product
	vars := mux.Vars(r)
	slug := vars["slug"]

	for _, p := range products {
		if p.Slug == slug {
			product = p
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if product.Slug != "" {
		payload, _ := json.Marshal(product)
		w.Write([]byte(payload))
	} else {
		w.Write([]byte("Product Not Found"))
	}
})

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := []byte("f20oKj9uwuX5QoREFVA6nDLtMwnEbUq0")
		secretProvider := auth0.NewKeyProvider(secret)
		audience := []string{"wervr"}

		configuration := auth0.NewConfiguration(secretProvider, audience, "https://dev-e6mv1-rq.auth0.com/", jose.HS256)
		validator := auth0.NewValidator(configuration, nil)

		token, err := validator.ValidateRequest(r)
		if err != nil {
			fmt.Println(err)
			fmt.Println("Token is not valid:", token)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

func main() {
	// instantiate gorilla/mux router
	r := mux.NewRouter()

	// Static index page
	r.Handle("/", http.FileServer(http.Dir("./views/")))

	// Serve static assets like css, images from /static/{file} route
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))

	// API will consist of two routes:
	// /products - which will retrieve a list of products that the user can leave feedback on
	// /products/{slug}/feedback - which will capture user feedback on products
	r.Handle("/products", authMiddleware(ProductsHandler)).Methods("GET")
	r.Handle("/products/{slug}/feedback", authMiddleware(AddFeedbackHandler)).Methods("POST")

	// Declare port 3000 and pass in router
	// Wrap the LoggingHandler function around our router so that the logger is called first on each route request
	http.ListenAndServe(":3000", handlers.LoggingHandler(os.Stdout, r))
}
