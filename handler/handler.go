package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	gcontext "github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/nirasan/gae-mobile-backend/bindata"
	"github.com/satori/go.uuid"
	"github.com/urfave/negroni"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

func NewHandler() http.Handler {
	r := mux.NewRouter()

	public := mux.NewRouter().PathPrefix("/user").Subrouter()
	public.HandleFunc("/registration", RegistrationHandler)
	public.HandleFunc("/authentication", AuthenticationHandler)

	r.PathPrefix("/user").Handler(public)

	auth := mux.NewRouter().PathPrefix("/").Subrouter()
	auth.HandleFunc("/hello", HelloWorldHandler)

	r.PathPrefix("/").Handler(negroni.New(
		negroni.HandlerFunc(AuthorizationMiddleware),
		negroni.Wrap(auth),
	))

	return r
}

type RegistrationHandlerRequest struct {
}

type RegistrationHandlerResponse struct {
	Success   bool
	UserID    int64
	UserToken string
}

type UserData struct {
	IntID     int64
	UserToken string
}

const (
	registrationRetryMax = 3
	userDataStoreName    = "UserData"
	userDataContextKey   = "UserData"
)

func RegistrationHandler(w http.ResponseWriter, r *http.Request) {

	var req RegistrationHandlerRequest
	DecodeJson(r, &req)

	ctx := appengine.NewContext(r)

	var userData UserData

	UserToken := uuid.NewV4().String()

	key := datastore.NewIncompleteKey(ctx, userDataStoreName, nil)
	userData = UserData{UserToken: UserToken}

	var err error
	if key, err = datastore.Put(ctx, key, &userData); err != nil {
		log.Errorf(ctx, "Failed to registration: %v", req)
		EncodeJson(w, RegistrationHandlerResponse{Success: false})
		return
	}

	// denormalization metadata
	userData.IntID = key.IntID()
	if _, err := datastore.Put(ctx, key, &userData); err != nil {
		log.Errorf(ctx, "Faild to registration: %v (%v)", req, err)
		EncodeJson(w, RegistrationHandlerResponse{Success: false})
		return
	}

	EncodeJson(w, RegistrationHandlerResponse{
		Success: true,
		UserID: userData.IntID,
		UserToken: userData.UserToken,
	})
}

type AuthenticationHandlerRequest struct {
	UserID int64
	UserToken string
}

type AuthenticationHandlerResponse struct {
	Success     bool
	AccessToken string
}

func AuthenticationHandler(w http.ResponseWriter, r *http.Request) {
	var req AuthenticationHandlerRequest
	DecodeJson(r, &req)

	ctx := appengine.NewContext(r)

	query := datastore.NewQuery(userDataStoreName).KeysOnly().Filter("IntID =", req.UserID).Filter("UserToken =", req.UserToken)
	keys, err := query.GetAll(ctx, nil)
	if err != nil || len(keys) != 1 {
		log.Errorf(ctx, "User not found: %v (%v)", req, err)
		EncodeJson(w, AuthenticationHandlerResponse{Success: false})
		return
	}

	method := jwt.GetSigningMethod("ES256")
	UserToken := jwt.NewWithClaims(method, jwt.MapClaims{
		"sub": keys[0].IntID(),
		"exp": time.Now().Add(time.Hour * 1).Unix(),
	})
	pem, e := bindata.Asset("ec256-key-pri.pem")
	if e != nil {
		panic(e.Error())
	}
	privateKey, e := jwt.ParseECPrivateKeyFromPEM(pem)
	if e != nil {
		panic(e.Error())
	}
	signedUserToken, e := UserToken.SignedString(privateKey)
	if e != nil {
		panic(e.Error())
	}
	EncodeJson(w, AuthenticationHandlerResponse{Success: true, AccessToken: signedUserToken})
}

type HelloWorldHandlerResponse struct {
	Success bool
	Message string
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {

	// Get UserData from gorilla context.
	userData, ok := gcontext.GetOk(r, userDataContextKey)
	if !ok {
		EncodeJson(w, HelloWorldHandlerResponse{Success: false})
		return
	}

	EncodeJson(w, HelloWorldHandlerResponse{Success: true, Message: fmt.Sprintf("Hello %d", userData.(UserData).IntID)})
}

func AuthorizationMiddleware(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {

	ctx := appengine.NewContext(r)

	// Get token from header
	header := r.Header.Get("Authorization")
	if header == "" {
		log.Errorf(ctx, "Invalid authorization hader")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	parts := strings.SplitN(header, " ", 2)
	if !(len(parts) == 2 && parts[0] == "Bearer") {
		log.Errorf(ctx, "Invalid authorization hader")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Check token
	token, e := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
		method := jwt.GetSigningMethod("ES256")
		if method != t.Method {
			return nil, errors.New("Invalid signing method")
		}
		pem, e := bindata.Asset("ec256-key-pub.pem")
		if e != nil {
			return nil, e
		}
		key, e := jwt.ParseECPublicKeyFromPEM(pem)
		if e != nil {
			return nil, e
		}
		return key, nil
	})
	if e != nil {
		log.Errorf(ctx, e.Error())
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get UserData from token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		log.Errorf(ctx, "invalid token: %v, %v", ok, token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	key := datastore.NewKey(ctx, userDataStoreName, "", int64(claims["sub"].(float64)), nil)
	var userData UserData
	if e := datastore.Get(ctx, key, &userData); e != nil {
		log.Errorf(ctx, "user not found: %v", e)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Set UserData to gorilla context
	gcontext.Set(r, userDataContextKey, userData)

	next(w, r)
}

func DecodeJson(r *http.Request, data interface{}) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	if e := decoder.Decode(data); e != nil {
		panic(e.Error())
	}
}

func EncodeJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
