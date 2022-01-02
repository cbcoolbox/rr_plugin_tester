package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/spiral/roadrunner-plugins/v2/http/attributes"
	jose "gopkg.in/square/go-jose.v2"
	jwt "gopkg.in/square/go-jose.v2/jwt"
)

// PluginName contains default service name.
const PluginName = "auth"
const cookieLength = 4 * time.Hour
const tokenLength = 30 * time.Second

var tempSigningKey = []byte("lol")
var tempEncKey = []byte("itsa16bytesecret")

// needed stuff for RR plugin
type Plugin struct {
	// server configuration (location, forbidden files and etc)
}

func (s *Plugin) Init() error {
	return nil
}

func (s *Plugin) Name() string {
	return PluginName
}

// Claim to make it easier with jwt
type CustomClaims struct {
	*jwt.Claims
	Tableit map[string]interface{} `json:"tableit"`
}

// Middleware must return true if request/response pair is handled within the middleware.
func (s *Plugin) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// check if we got a lil cookie
		token, err := r.Cookie("tableitAuth")

		if err != nil {
			if err.Error() != "http: named cookie not present" {
				handleErr(err.Error(), w, r, next)
				return
			}
		}

		if token != nil {
			// verify jwt
			parsedToken, err := jwt.ParseSignedAndEncrypted(token.Value)

			// if we cant parse the key then we give em a new one
			if err != nil {
				handleNewAuthCookie(w, r, next)
				return
			}

			// decypt the jw(e?)
			nested, err := parsedToken.Decrypt(tempEncKey)
			if err != nil {
				handleNewAuthCookie(w, r, next)
				return
			}

			out := CustomClaims{}

			// if the claims dont match with the key we give em a new one
			if err := nested.Claims(tempSigningKey, &out); err != nil {
				handleNewAuthCookie(w, r, next)
				return
			}

			// wow gj but now we have to validate expirary
			err = out.Validate(jwt.Expected{
				Time: time.Now(),
			})

			if err != nil {
				handleNewAuthCookie(w, r, next)
				return
			}

			if err != nil {
				w.Header().Add("error", err.Error())
			}

			// if you made it here good job! you actually sent us a real token!
			r = attributes.Init(r)
			attributes.Set(r, "auth", out.Tableit)
			w.Header().Add("authed", "yaya")
			next.ServeHTTP(w, r)
		} else {
			// no token, give out a newbie
			handleNewAuthCookie(w, r, next)
		}
	})
}

func handleNewAuthCookie(w http.ResponseWriter, r *http.Request, next http.Handler) {
	token, auth, err := makeJWT()

	if err != nil {
		handleErr(err.Error(), w, r, next)
		return
	}

	r = attributes.Init(r)
	attributes.Set(r, "auth", auth)

	addCookie(w, "tableitAuth", token, cookieLength)

	next.ServeHTTP(w, r)
}

func makeJWT() (string, string, error) {
	// make the signer
	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.HS512,
			Key:       tempSigningKey,
		}, (&jose.SignerOptions{}).WithType("JWT"))

	if err != nil {
		return "", "", err
	}

	// make the encrypter
	encrypter, err := jose.NewEncrypter(
		jose.A128GCM,
		jose.Recipient{
			Algorithm: jose.DIRECT,
			Key:       tempEncKey,
		},
		(&jose.EncrypterOptions{}).WithType("JWT").WithContentType("JWT"))

	if err != nil {
		return "", "", err
	}

	cc := CustomClaims{
		Claims: &jwt.Claims{
			Issuer: "tableit",
			Expiry: jwt.NewNumericDate(time.Now().Add(tokenLength)),
		},
		Tableit: map[string]interface{}{
			"security_id":    0,
			"application_id": 1,
		},
	}

	raw, err := jwt.SignedAndEncrypted(signer, encrypter).Claims(cc).CompactSerialize()

	if err != nil {
		return "", "", err
	}

	auth, _ := json.Marshal(cc.Tableit)

	return raw, string(auth), nil
}

func handleErr(err string, w http.ResponseWriter, r *http.Request, next http.Handler) {
	w.Header().Add("error", err)
	next.ServeHTTP(w, r)
}

func addCookie(w http.ResponseWriter, name, value string, ttl time.Duration) {
	expire := time.Now().Add(ttl)
	cookie := http.Cookie{
		Name:    name,
		Value:   value,
		Expires: expire,
	}
	http.SetCookie(w, &cookie)
}
