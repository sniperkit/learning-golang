package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/context"
	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Repo

type Tea struct {
	Id       bson.ObjectId `json:"id,omitempty" bson:"_id,omitempty"`
	Name     string        `json:"name"`
	Category string        `json:"category"`
}

type TeaResource struct {
	Data Tea `json:"data"`
}

type TeaRepo struct {
	coll *mgo.Collection
}

func (r *TeaRepo) Find(id string) (TeaResource, error) {
	result := TeaResource{}
	err := r.coll.FindId(bson.ObjectIdHex(id)).One(&result.Data)
	if err != nil {
		return result, err
	}

	return result, nil
}

type appContext struct {
	db *mgo.Database
}

func (c *appContext) teaHandler(w http.ResponseWriter, r *http.Request) {
	params := context.Get(r, "params").(httprouter.Params)
	repo := TeaRepo{c.db.C("teas")}
	tea, err := repo.Find(params.ByName("id"))
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/vnd.api+json")
	json.NewEncoder(w).Encode(tea)
}

// func recoverHandler(next http.Handler) http.Handler {
// 	fn := func(w http.ResponseWriter, r *http.Request) {
// 		defer func() {
// 			if err := recover(); err != nil {
// 				log.Printf("panic: %+v", err)
// 				http.Error(w, http.StatusText(500), 500)
// 			}
// 		}()

// 		next.ServeHTTP(w, r)
// 	}

// 	return http.HandlerFunc(fn)
// }

type Errors struct {
	// We use an array of *Error to set common errors in variables.
	// More on that later.
	Errors []*Error `json:"errors"`
}

type Error struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func recoverHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %+v", err)
				// http.Error(w, http.StatusText(500), 500)
				jsonErr := &Error{"internal_server_error", 500, "Internal Server Error", "Something went wrong."}
				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(jsonErr.Status)
				json.NewEncoder(w).Encode(Errors{[]*Error{jsonErr}})
			}
		}()

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func loggingHandler(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[%s] %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	}

	return http.HandlerFunc(fn)
}

type router struct {
	*httprouter.Router
}

func (r *router) Get(path string, handler http.Handler) {
	r.GET(path, wrapHandler(handler))
}

func NewRouter() *router {
	return &router{httprouter.New()}
}

func wrapHandler(h http.Handler) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		context.Set(r, "params", ps)
		h.ServeHTTP(w, r)
	}
}

func main() {
	session, err := mgo.Dial("10.0.6.22")
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	appC := appContext{session.DB("test")}
	commonHandlers := alice.New(context.ClearHandler, loggingHandler, recoverHandler)
	router := NewRouter()
	router.Get("/teas/:id", commonHandlers.ThenFunc(appC.teaHandler))
	http.ListenAndServe(":8888", router)
}
