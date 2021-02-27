package proxy

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	jwt "github.com/dgrijalva/jwt-go/v4"
)

func handleError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"error\": \"%s\"}", err)
}

func validateRequest(httpMethod string, httpPath string, allowedAPIMethods string, k8sAllowedAPIRegexp *regexp.Regexp) error {
	// Validate method
	if allowedAPIMethods != "" {
		if !strings.Contains(allowedAPIMethods, strings.ToLower(httpMethod)) {
			return fmt.Errorf("%s method not allowedd", httpMethod)
		}
	}

	// Validate path
	if !k8sAllowedAPIRegexp.MatchString(httpPath) && httpPath != "/.well-known/oauth-authorization-server" {
		return fmt.Errorf("%s path not allowed", httpPath)
	}

	return nil
}

func validateToken(token string, key []byte, httpMethod string, httpPath string) (*jwt.Token, error) {
	tok, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		return key, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
		var allowedAPIMethods string
		var allowedAPIRegexp string

		if allowedAPIMethods, ok = claims["allowedAPIMethods"].(string); !ok {
			allowedAPIMethods = ""
		}
		if allowedAPIRegexp, ok = claims["allowedAPIRegexp"].(string); !ok {
			allowedAPIRegexp = ""
		}
		k8sAllowedAPIRegexp := regexp.MustCompile(allowedAPIRegexp)

		err := validateRequest(httpMethod, httpPath, allowedAPIMethods, k8sAllowedAPIRegexp)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("invalid bearer token")
	}

	return tok, nil
}
