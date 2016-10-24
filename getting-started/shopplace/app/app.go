package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"cloud.google.com/go/storage"

	"golang.org/x/net/context"

	"github.com/Chidi150/golang-samples/getting-started/shopplace"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/satori/go.uuid"
	"google.golang.org/appengine"
)

var (
	// See template.go
	firstTmpl  = parseTemplate("first.html")
	listTmpl   = parseTemplate("list.html")
	editTmpl   = parseTemplate("edit.html")
	detailTmpl = parseTemplate("detail.html")
)

func main() {
	registerHandlers()
	appengine.Main()
}

func registerHandlers() {
	// Use gorilla/mux for rich routing.
	// See http://www.gorillatoolkit.org/pkg/mux
	r := mux.NewRouter()
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	//r.Handle("/", http.RedirectHandler("/shops", http.StatusFound))

	r.Methods("GET").Path("/").
		Handler(appHandler(firstHandler))
	r.Methods("GET").Path("/shops").
		Handler(appHandler(listHandler))
	r.Methods("GET").Path("/shops/category/{name:[A-Z' a-z]+}").
		Handler(appHandler(listCategoryHandler))
	r.Methods("GET").Path("/shops/mine").
		Handler(appHandler(listMineHandler))
	r.Methods("GET").Path("/shops/{id:[0-9]+}").
		Handler(appHandler(detailHandler))
	r.Methods("GET").Path("/shops/add").
		Handler(appHandler(addFormHandler))
	r.Methods("GET").Path("/shops/{id:[0-9]+}/edit").
		Handler(appHandler(editFormHandler))
	r.Methods("GET").Path("/signup").
		Handler(appHandler(showSignupFormHandler))

	r.Methods("POST").Path("/shops").
		Handler(appHandler(createHandler))
	r.Methods("POST", "PUT").Path("/shops/{id:[0-9]+}").
		Handler(appHandler(updateHandler))
	r.Methods("POST").Path("/shops/{id:[0-9]+}:delete").
		Handler(appHandler(deleteHandler)).Name("delete")

	// The following handlers are defined in auth.go and used in the
	// "Authenticating Users" part of the Getting Started guide.

	r.Methods("GET").Path("/login").
		Handler(appHandler(loginHandler))
	r.Methods("POST").Path("/logout").
		Handler(appHandler(logoutHandler))
	r.Methods("GET").Path("/oauth2callback").
		Handler(appHandler(oauthCallbackHandler))

	// Respond to App Engine and Compute Engine health checks.
	// Indicate the server is healthy.
	r.Methods("GET").Path("/_ah/health").HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("ok"))
		})

	// [START request_logging]
	// Delegate all of the HTTP routing and serving to the gorilla/mux router.
	// Log all requests using the standard Apache format.
	http.Handle("/", handlers.CombinedLoggingHandler(os.Stderr, r))
	// [END request_logging]
}

// listHandler displays a list with summaries of shops in the database.
func firstHandler(w http.ResponseWriter, r *http.Request) *appError {
	//	shops, err := shopplace.DB.ListShops()
	//	if err != nil {
	//		return appErrorf(err, "could not list shops: %v", err)
	//	}

	return firstTmpl.Execute(w, r, nil)
}

// listHandler displays a list with summaries of shops in the database.
func listHandler(w http.ResponseWriter, r *http.Request) *appError {
	shops, err := shopplace.DB.ListShops()
	if err != nil {
		return appErrorf(err, "could not list shops: %v", err)
	}

	return listTmpl.Execute(w, r, shops)
}

// listCategoryHandler displays a list of shops of a particular category
func listCategoryHandler(w http.ResponseWriter, r *http.Request) *appError {
	params := mux.Vars(r)
	category := params["name"]

	shops, err := shopplace.DB.ListShopsCategory(category)
	if err != nil {

		return appErrorf(err, "could not list shops: %v", err)
	}

	return listTmpl.Execute(w, r, shops)
}

// listMineHandler displays a list of shops created by the currently
// authenticated user.
func listMineHandler(w http.ResponseWriter, r *http.Request) *appError {
	user := profileFromSession(r)
	if user == nil {
		http.Redirect(w, r, "/login?redirect=/shops/mine", http.StatusFound)

		return nil
	}

	shops, err := shopplace.DB.ListShopsCreatedBy(user.Id)
	if err != nil {

		return appErrorf(err, "could not list shops: %v", err)
	}

	return listTmpl.Execute(w, r, shops)
}

// shopFromRequest retrieves a shop from the database given a shop ID in the
// URL's path.
func shopFromRequest(r *http.Request) (*shopplace.Shop, error) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("bad shop id: %v", err)
	}
	shop, err := shopplace.DB.GetShop(id)
	if err != nil {
		return nil, fmt.Errorf("could not find shop: %v", err)
	}
	return shop, nil
}

// detailHandler displays the details of a given shop.
func detailHandler(w http.ResponseWriter, r *http.Request) *appError {
	shop, err := shopFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return detailTmpl.Execute(w, r, shop)
}

// addFormHandler displays a form that captures details of a new shop to add to
// the database.
func addFormHandler(w http.ResponseWriter, r *http.Request) *appError {

	return editTmpl.Execute(w, r, nil)

}

// ShowSignupFormHandler displays a form that captures details of a new user to add to
// the database.
func showSignupFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	return editTmpl.Execute(w, r, nil)
}

// editFormHandler displays a form that allows the user to edit the details of
// a given shop.
func editFormHandler(w http.ResponseWriter, r *http.Request) *appError {
	shop, err := shopFromRequest(r)
	if err != nil {
		return appErrorf(err, "%v", err)
	}

	return editTmpl.Execute(w, r, shop)
}

// shopFromForm populates the fields of a Shop from form values
// (see templates/edit.html).
func shopFromForm(r *http.Request) (*shopplace.Shop, error) {
	imageURL, err := uploadFileFromForm(r)
	if err != nil {
		return nil, fmt.Errorf("could not upload file: %v", err)
	}
	if imageURL == "" {
		imageURL = r.FormValue("imageURL")
	}

	shop := &shopplace.Shop{
		Title:         r.FormValue("title"),
		Author:        r.FormValue("author"),
		PublishedDate: r.FormValue("publishedDate"),
		ImageURL:      imageURL,
		Description:   r.FormValue("description"),
		CreatedBy:     r.FormValue("createdBy"),
		CreatedByID:   r.FormValue("createdByID"),
		Category:      r.FormValue("category"),
		Address:       r.FormValue("address"),
		EmailAddress:  r.FormValue("emailaddress"),
		Phone:         r.FormValue("phone"),
	}

	// If the form didn't carry the user information for the creator, populate it
	// from the currently logged in user (or mark as anonymous).
	if shop.CreatedByID == "" {
		user := profileFromSession(r)
		if user != nil {
			// Logged in.
			shop.CreatedBy = user.DisplayName
			shop.CreatedByID = user.Id
		} else {
			// Not logged in.
			shop.SetCreatorAnonymous()
		}
	}

	return shop, nil
}

// uploadFileFromForm uploads a file if it's present in the "image" form field.
func uploadFileFromForm(r *http.Request) (url string, err error) {
	f, fh, err := r.FormFile("image")
	if err == http.ErrMissingFile {
		return "", nil
	}
	if err != nil {
		return "", err
	}

	if shopplace.StorageBucket == nil {
		return "", errors.New("storage bucket is missing - check config.go")
	}

	// random filename, retaining existing extension.
	name := uuid.NewV4().String() + path.Ext(fh.Filename)

	ctx := context.Background()
	w := shopplace.StorageBucket.Object(name).NewWriter(ctx)
	w.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}
	w.ContentType = fh.Header.Get("Content-Type")

	// Entries are immutable, be aggressive about caching (1 day).
	w.CacheControl = "public, max-age=86400"

	if _, err := io.Copy(w, f); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	const publicURL = "https://storage.googleapis.com/%s/%s"
	return fmt.Sprintf(publicURL, shopplace.StorageBucketName, name), nil
}

// createHandler adds a shop to the database.
func createHandler(w http.ResponseWriter, r *http.Request) *appError {
	shop, err := shopFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse shop from form: %v", err)
	}
	id, err := shopplace.DB.AddShop(shop)
	if err != nil {
		return appErrorf(err, "could not save shop: %v", err)
	}
	http.Redirect(w, r, fmt.Sprintf("/shops/%d", id), http.StatusFound)
	return nil
}

// updateHandler updates the details of a given shop.
func updateHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad shop id: %v", err)
	}

	shop, err := shopFromForm(r)
	if err != nil {
		return appErrorf(err, "could not parse shop from form: %v", err)
	}
	shop.ID = id

	err = shopplace.DB.UpdateShop(shop)
	if err != nil {
		return appErrorf(err, "could not save shop: %v", err)
	}

	http.Redirect(w, r, fmt.Sprintf("/shops/%d", shop.ID), http.StatusFound)
	return nil
}

// deleteHandler deletes a given shop.
func deleteHandler(w http.ResponseWriter, r *http.Request) *appError {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		return appErrorf(err, "bad shop id: %v", err)
	}
	err = shopplace.DB.DeleteShop(id)
	if err != nil {
		return appErrorf(err, "could not delete shop: %v", err)
	}
	http.Redirect(w, r, "/shops", http.StatusFound)
	return nil
}

// http://blog.golang.org/error-handling-and-go
type appHandler func(http.ResponseWriter, *http.Request) *appError

type appError struct {
	Error   error
	Message string
	Code    int
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e := fn(w, r); e != nil { // e is *appError, not os.Error.
		log.Printf("Handler error: status code: %d, message: %s, underlying err: %#v",
			e.Code, e.Message, e.Error)

		http.Error(w, e.Message, e.Code)
	}
}

func appErrorf(err error, format string, v ...interface{}) *appError {
	return &appError{
		Error:   err,
		Message: fmt.Sprintf(format, v...),
		Code:    500,
	}
}
