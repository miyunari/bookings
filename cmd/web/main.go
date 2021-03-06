package main

import (
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MninaTB/bookings/internal/config"
	"github.com/MninaTB/bookings/internal/handlers"
	"github.com/MninaTB/bookings/internal/models"
	"github.com/MninaTB/bookings/internal/render"
	"github.com/alexedwards/scs/v2"
)

const portNumber = ":8080"

var app config.AppConfig
var session *scs.SessionManager

// main is the entry point of the application
func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(fmt.Sprintf("Starting application on port %s", portNumber))

	srv := &http.Server{
		Addr:    portNumber,
		Handler: routes(&app),
	}
	err = srv.ListenAndServe()
	log.Fatal(err)
}

func run() error {
	// what am I going to put in the session
	gob.Register(models.Reservation{})

	// change this to true when in production
	app.InProduction = false

	session = scs.New()
	// the time the session will last
	session.Lifetime = 24 * time.Hour
	// the cookies should persist, after the user closes the browser window
	session.Cookie.Persist = true
	session.Cookie.SameSite = http.SameSiteLaxMode
	// should the cookie be encrypted - only works with https
	session.Cookie.Secure = app.InProduction

	app.Session = session

	tc, err := render.CreateTemplateCache()
	if err != nil {
		log.Fatal("cannot create Template Cache")
		return err
	}

	app.TemplateCache = tc
	app.UseCache = false

	repo := handlers.NewRepo(&app)
	handlers.NewHandlers(repo)

	render.NewTemplates(&app)

	return nil
}
