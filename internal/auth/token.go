/******************************************************************************
*
*  Copyright 2018-2019 SAP SE
*
*  Licensed under the Apache License, Version 2.0 (the "License");
*  you may not use this file except in compliance with the License.
*  You may obtain a copy of the License at
*
*      http://www.apache.org/licenses/LICENSE-2.0
*
*  Unless required by applicable law or agreed to in writing, software
*  distributed under the License is distributed on an "AS IS" BASIS,
*  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*  See the License for the specific language governing permissions and
*  limitations under the License.
*
******************************************************************************/

package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/docker/libtrust"
	"github.com/sapcc/keppel/internal/keppel"
)

//Token represents a JWT (Java Web Token), as used for authenticating on the
//Registry v2 API.
type Token struct {
	//The name of the user who created this token. For anonymous users, this is empty.
	UserName string
	//The service that this token can be used with.
	Audience string
	//Access permissions for this token.
	Access []Scope
}

//TokenClaims is the type for JWT claims issued by Keppel.
type TokenClaims struct {
	jwt.StandardClaims
	Access []Scope `json:"access"`
}

//ParseTokenFromRequest tries to parse the Bearer token supplied in the
//request's Authorization header.
func ParseTokenFromRequest(r *http.Request, cfg keppel.Configuration) (*Token, *keppel.RegistryV2Error) {
	//read Authorization request header
	tokenStr := r.Header.Get("Authorization")
	if !strings.HasPrefix(tokenStr, "Bearer ") { //e.g. because it's missing
		return nil, keppel.ErrUnauthorized.With("no bearer token found in request headers")
	}
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	//parse JWT
	var claims TokenClaims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (interface{}, error) {
		//check that the signing method matches what we generate
		ourIssuerKey := cfg.JWTIssuerKey
		ourSigningMethod := ChooseSigningMethod(ourIssuerKey)
		if !equalSigningMethods(ourSigningMethod, t.Method) {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		//jwt.Parse needs the public key for our issuer key to validate the token
		return ourIssuerKey.PublicKey().CryptoPublicKey(), nil
	})
	if err != nil {
		return nil, keppel.ErrUnauthorized.With(err.Error())
	}
	if !token.Valid {
		//NOTE: This branch is defense in depth. As of the time of this writing,
		//token.Valid == false if and only if err != nil.
		return nil, keppel.ErrUnauthorized.With("token invalid")
	}

	//check claims (allow up to 3 seconds clock mismatch)
	now := time.Now().Unix()
	if !claims.StandardClaims.VerifyExpiresAt(now-3, true) {
		return nil, keppel.ErrUnauthorized.With("token expired")
	}
	if !claims.StandardClaims.VerifyNotBefore(now+3, true) {
		return nil, keppel.ErrUnauthorized.With("token not valid yet")
	}
	publicHost := cfg.APIPublicHostname()
	if !claims.StandardClaims.VerifyIssuer("keppel-api@"+publicHost, true) {
		return nil, keppel.ErrUnauthorized.With("token has wrong issuer")
	}
	if !claims.StandardClaims.VerifyAudience(publicHost, true) {
		return nil, keppel.ErrUnauthorized.With("token has wrong audience")
	}

	return &Token{
		UserName: claims.StandardClaims.Subject,
		Audience: publicHost,
		Access:   claims.Access,
	}, nil
}

//Contains returns true if the given token authorizes the user for this scope.
func (t Token) Contains(s Scope) bool {
	for _, scope := range t.Access {
		if scope.Contains(s) {
			return true
		}
	}
	return false
}

//ChooseSigningMethod returns the appropriate signing method for the given
//private key.
func ChooseSigningMethod(key libtrust.PrivateKey) jwt.SigningMethod {
	issuerKey := key.CryptoPrivateKey()
	switch issuerKey.(type) {
	case *ecdsa.PrivateKey:
		return jwt.SigningMethodES256
	case *rsa.PrivateKey:
		return jwt.SigningMethodRS256
	default:
		panic(fmt.Sprintf("do not know which JWT method to use for issuerKey.type = %T", issuerKey))
	}
}

func equalSigningMethods(m1, m2 jwt.SigningMethod) bool {
	switch m1 := m1.(type) {
	case *jwt.SigningMethodECDSA:
		if m2, ok := m2.(*jwt.SigningMethodECDSA); ok {
			return *m1 == *m2
		}
		return false
	case *jwt.SigningMethodRSA:
		if m2, ok := m2.(*jwt.SigningMethodRSA); ok {
			return *m1 == *m2
		}
		return false
	default:
		panic(fmt.Sprintf("do not know how to compare signing methods of type %T", m1))
	}
}
