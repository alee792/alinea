package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	pb "github.com/alee792/alinea/proto"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var log = logrus.New()

// Config pulls in environment variables
type Config struct {
	Port        string `default:"8080"`
	DbUser      string `default:"pushygo"`
	DbPassword  string `default:"pushygo"`
	DbHost      string `default:"pushygo-db"`
	DbPort      int    `default:"5432"`
	DbName      string `default:"pushygo"`
	Environment string `default:"dev"`
}

// App describes the PushyGo web client
type App struct {
	Router *mux.Router
	DB     *sql.DB
}

// PushFormInput holds info relevant to a URL request
type PushFormInput struct {
	DeviceID string `json:"deviceID"`
	Content  pb.Content
}

func main() {
	a := App{}
	var c Config
	err := envconfig.Process("pg", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	if c.Environment == "dev" {
		log.Level = logrus.DebugLevel
	}
	a.DB = newDB(&c)
	a.Router = mux.NewRouter()
	a.Router.HandleFunc("/device", a.createDevice).Methods("POST")
	a.Router.HandleFunc("/device", a.getDashboard).Methods("GET")
	a.Router.HandleFunc("/push", a.pushContent).Methods("POST")
	a.Router.HandleFunc("/create", a.getCreateForm).Methods("Get")
	a.Router.HandleFunc("/", a.getHome).Methods("GET")

	log.Infof("Listening on port %s", c.Port)
	log.Fatal(http.ListenAndServe(":"+c.Port, a.Router))

}

func (a *App) createDevice(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	d := new(Device)
	decoder := schema.NewDecoder()
	err := decoder.Decode(d, r.PostForm)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	log.Debug(d)
	err = d.create(a.DB)
	if err != nil {
		log.Error(err)
		respondWithError(w, http.StatusBadRequest, "Unable to create device: "+err.Error())
		return
	}
	respondWithJSON(w, http.StatusCreated, d)
	return
}

func (a *App) getHome(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "home", "home", nil)
	return
}

func (a *App) getCreateForm(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "device", "create", nil)
	return
}
func (a *App) getDashboard(w http.ResponseWriter, r *http.Request) {
	ds, err := getAll(a.DB, 100, 0)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	actions := map[string]string{
		"http":    "/url",
		"https":   "/url",
		"youtube": "youtube",
	}
	dd := &DeviceData{
		PageTitle: "Devices",
		Devices:   ds,
		Actions:   actions,
	}
	renderTemplate(w, "device", "device", dd)
	return
}
func (a *App) pushContent(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	f := new(PushFormInput)
	decoder := schema.NewDecoder()
	err := decoder.Decode(f, r.PostForm)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	var d Device
	err = d.getDevice(a.DB, f.DeviceID)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	conn, err := dialService(d.URL)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	client := pb.NewContentPushClient(conn)
	client.PushContent(nil, &f.Content, nil)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
}

func dialService(address string) (conn *grpc.ClientConn, err error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithBackoffMaxDelay(1*time.Second))
	conn, err = grpc.Dial(address, opts...)
	if err != nil {
		return
	}
	return
}

func newDB(c *Config) (db *sql.DB) {
	dbURI := fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable",
		c.DbUser, c.DbPassword, c.DbHost, c.DbName,
	)
	log.Debug(dbURI)
	db, err := sql.Open("postgres", dbURI)
	if err != nil {
		panic(err)
	}
	return
}
