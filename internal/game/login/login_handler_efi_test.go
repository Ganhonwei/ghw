package login

import (
	"testing"
)

func TestEfiGetGameList(t *testing.T) {
	resp, err := EfiGetGameList("100102", "joker", "185.14.47.96")
	if err != nil {
		t.Errorf("%+v", err)
		t.Error(err.Error())
		return
	}
	t.Log(resp)
}

func TestEfiLaunchGame(t *testing.T) {
	resp, err := EfiLaunchGame("100707", "joker", "185.14.47.96", "rng-topcard00001") // rkfaxspyyoeqae3g
	if err != nil {
		t.Log(err)
		t.Error(err)
	}
	t.Log(resp)
}

func TestEfi3(t *testing.T) {

}
