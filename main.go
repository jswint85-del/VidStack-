package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
)

const domain = "https://vidstack-4.polsia.app/pricing"

var plans = map[string]string{
	"starter":  "price_1TfRFdPxwjKtw7b1OGnUYtgw",
	"operator": "price_1TfRHCPxwjKtw7b17Lb3nQKI",
	"studio":   "price_1TfRLuPxwjKtw7b1ZTrezLK0",
}

func main() {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	if stripe.Key == "" {
		log.Fatal("STRIPE_SECRET_KEY environment variable is required")
	}

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/create-checkout-session", checkoutHandler)
	http.HandleFunc("/success", successHandler)
	http.HandleFunc("/cancel", cancelHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var body struct {
		Plan string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	priceID, ok := plans[body.Plan]
	if !ok {
		http.Error(w, "Invalid plan", http.StatusBadRequest)
		return
	}

	params := &stripe.CheckoutSessionParams{
		Mode: stripe.String(string(stripe.CheckoutSessionModePayment)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL: stripe.String(domain + "/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:  stripe.String(domain + "/cancel"),
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("checkout session error: %v", err)
		http.Error(w, "Failed to create checkout session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"url": s.URL})
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/success.html"))
	tmpl.Execute(w, nil)
}

func cancelHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/cancel.html"))
	tmpl.Execute(w, nil)
}
