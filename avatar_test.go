package main

import (
	"testing"
)

func TestOpenIDAvatarMissingAvatarURLKey(t *testing.T) {
	var avatar OpenIDAvatar
	client := new(client)
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("OpenIDAvatar.GetAvatarURL should return ErrNoAvatarURL when AvatarURL is missing.")
	}
}

func TestOpenIDAvatarWrongType(t *testing.T) {
	var avatar OpenIDAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		avatarURLKey: 10,
	}
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("OpenIDAvatar.GetAvatarURL should return ErrNoAvatarURL when AvatarURL has the wrong type.")
	}
}

func TestOpenIDAvatarEmptyString(t *testing.T) {
	var avatar OpenIDAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		avatarURLKey: "",
	}
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("OpenIDAvatar.GetAvatarURL should return ErrNoAvatarURL when AvatarURL is an empty string.")
	}
}

func TestOpenIDAvatarValidURL(t *testing.T) {
	expAvatarURL := "https://foo.com/some/image.jpeg"
	var avatar OpenIDAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		avatarURLKey: expAvatarURL,
	}

	avatarURL, err := avatar.GetAvatarURL(client)
	if err != nil {
		t.Error("OpenIDAvatar.GetAvatarURL should return correct URL.")
	}

	if expAvatarURL != avatarURL {
		t.Error("OpenIDAvatar.GetAvatarURL should return correct URL.")
	}
}

func TestGravatarAvatarMissingEmailKey(t *testing.T) {
	var avatar GravatarAvatar
	client := new(client)
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("GravatarAvatar.GetAvatarURL should return ErrNoAvatarURL when Email is missing.")
	}
}

func TestGravatarAvatarWrongType(t *testing.T) {
	var avatar GravatarAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		emailKey: 10,
	}
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("GravatarAvatar.GetAvatarURL should return ErrNoAvatarURL when Email has the wrong type.")
	}
}

func TestGravatarAvatarEmptyEmail(t *testing.T) {
	var avatar GravatarAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		emailKey: "",
	}
	_, err := avatar.GetAvatarURL(client)
	if err != ErrNoAvatarURL {
		t.Error("GravatarAvatar.GetAvatarURL should return ErrNoAvatarURL when Email is an empty string.")
	}
}

func TestGravatarAvatarValidEmail(t *testing.T) {
	expAvatarURL := "https://www.gravatar.com/avatar/93942e96f5acd83e2e047ad8fe03114d"
	var avatar GravatarAvatar
	client := new(client)
	client.userData = map[string]interface{}{
		emailKey: "TEST@EMAIL.com",
	}

	avatarURL, err := avatar.GetAvatarURL(client)
	if err != nil {
		t.Error("GravatarAvatar.GetAvatarURL should return correct URL.")
	}

	if expAvatarURL != avatarURL {
		t.Error("GravatarAvatar.GetAvatarURL should return correct URL.")
	}
}
