package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type Person struct {
	Name  string `json:"given_name"`
	Email string `json:"email"`
}

var listRandomStates map[string]bool = make(map[string]bool)

func randomHex(n int) (string, error) {
	// ***** TASK #2: FIX RANDOM STATE *****
	return "00000", nil
}

func googleConfig() *oauth2.Config {
	// credentials should be obtained from https://console.developers.google.com
	config := &oauth2.Config{
		ClientID:     os.Getenv("CLIENTID"),
		ClientSecret: os.Getenv("CLIENTSECRET"),
		RedirectURL:  "http://localhost:8080/callback",
		Scopes: []string{
			"openid",
			"email",
			"profile",
		},
		Endpoint: google.Endpoint,
	}
	return config
}

func handlerLogin(w http.ResponseWriter, req *http.Request) {
	fmt.Println("User is trying to login...")
	config := googleConfig()

	// redirect user to Google's consent page to ask for permission for the scopes specified in config
	// sample random state and add to listRandomStates
	state, _ := randomHex(32)
	listRandomStates[state] = true
	url := config.AuthCodeURL(state)
	fmt.Printf("User %v is being redirected to... %v\n", state, url)
	http.Redirect(w, req, url, http.StatusFound)
}

func handlerCallback(w http.ResponseWriter, req *http.Request) {
	// parse state
	state := req.FormValue("state")
	fmt.Printf("Callback from... %v\n", state)
	if listRandomStates[state] == true {
		// all good, mark state as used
		listRandomStates[state] = false
	} else {
		log.Fatal("Callback from whom? ¯\\_(ツ)_/¯")
	}

	// parse code
	code := req.FormValue("code")

	// we need config usable inside func handlerCallback
	config := googleConfig()

	// exchange authorization code for token
	token, _ := config.Exchange(oauth2.NoContext, code)

	// extract token_id
	idToken := token.Extra("id_token")
	fmt.Println(idToken)

	// access token is not needed for authentication but could be used to access resources
	//accessToken := token.Extra("access_token")

	// validation of idToken requires several steps:
	// (1) verify the signature against the issuer, see jwks_uri metadata value of the Discovery https://accounts.google.com/.well-known/openid-configuration;
	// (2) Verify that the value of the iss claim in the ID token is equal to https://accounts.google.com or accounts.google.com;
	// (3) verify that the value of the aud claim in the ID token is equal to your app's client ID;
	// (4) verify that the expiry time (exp claim) of the ID token has not passed;

	// ***** TASK #3: VALIDATE THE DIGITAL SIGNATURE OF 'idToken' *****
	// ***** HINT 1: Inspect 'idToken' and notice that 'kid' is the reference of Google's public key used to sign 'idToken' *****
	// ***** HINT 2: Import package "github.com/MicahParks/keyfunc" to select the correct public key based on the 'kid' of 'idToken' *****

	// parse the JWT idToken
	var myDummyKeyFunc jwt.Keyfunc = func(token *jwt.Token) (interface{}, error) {
		// this should actually return the public key to validate 'token'
		return nil, nil
	}
	parsedToken, err := jwt.Parse(idToken.(string), myDummyKeyFunc)

	// life is short and this is a demo, so I only check (1)
	if err != nil {
		fmt.Println("ERROR: Could not validate token!")
	} else {
		fmt.Println("Received a valid token!")
	}

	// parse claims and create person
	claims, _ := parsedToken.Claims.(jwt.MapClaims)
	person := Person{Name: claims["given_name"].(string), Email: claims["email"].(string)}
	fmt.Printf("Name: %v, Email: %v\n", person.Name, person.Email)

	// use template package to inject person into callback.html
	t, _ := template.ParseFiles("html/callback.html")
	t.Execute(w, person)
}

func main() {

	// ***** TASK #1: CREATE .env FILE AND DEFINE ENVIRONMENT VARIABLES 'CLIENTID' and 'CLIENTSECRET' *****

	// load .env file from given path
	// we keep it empty it will load .env from current directory
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	// handler
	http.Handle("/", http.FileServer(http.Dir("html")))
	http.HandleFunc("/login", handlerLogin)
	http.HandleFunc("/callback", handlerCallback)

	// start server
	fmt.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
