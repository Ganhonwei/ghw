package handler

import (
	"fmt"
	"goserver/pkg/utils"

	browser "github.com/EDDYCJY/fake-useragent"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pkg/errors"
)

var key = []byte{'j', 'x', 'z', 'b', 't', 'e', 'c', 'h'}

func Sign(uid string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": uid,
		// "exp": time.Now().Add(24 * time.Hour).Unix(),
	})

	token, err := t.SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

func SignByTime(uid string) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid":       uid,
		"timestamp": utils.BsonNow().UnixMilli(),
	})

	token, err := t.SignedString(key)
	if err != nil {
		return "", err
	}

	return token, nil
}

func Valid(token string) (string, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New(fmt.Sprintf("unexpected signing method: %v", token.Header["alg"]))
		}
		return key, nil
	})
	if err != nil {
		return "", err
	}

	if t.Valid {
		claims, ok := t.Claims.(jwt.MapClaims)
		if !ok {
			return "", errors.New("claims cannot be asserted to jwt.mapClaims")
		}

		uid, ok := claims["uid"].(string)
		if !ok {
			return "", errors.New("claims uid cannot be asserted to float64")
		}
		return uid, nil
	} else {
		return "", errors.New(fmt.Sprintf("token: %s is invalid", token))
	}
}

// 获取ua
func GenUserAgent() string {
	return browser.Android()
}
