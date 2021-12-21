package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"strings"
)

// ErrNoAvatarURL is the error that is returned
// when the Avatar object is unable to provide a valid avatar URL.
var ErrNoAvatarURL = errors.New("unable to get avatar url")

// Avatar is an abstraction of a type capable
// of providing an avatar URL.
type Avatar interface {
	// GetAvatarURL return the avatar URL for the input client
	// or ErrNoAvatarURL if unable to provide a valid avatar URL.
	GetAvatarURL(c *client) (string, error)
}

type OpenIDAvatar struct{}

type GravatarAvatar struct{}

// UseOpenIDAvatar has the OpenIDAvatar type (zero-initialized).
// We can assign the UseOpenIDAvatar to any field requiring an Avatar type.
var UseOpenIDAvatar OpenIDAvatar

// UseGravatarAvatar has the GravatarAvatar type (zero-initialized).
// We can assign the UseGravatarAvatar to any field requiring an Avatar type.
var UseGravatarAvatar GravatarAvatar

func (OpenIDAvatar) GetAvatarURL(c *client) (string, error) {
	v, ok := c.userData[avatarURLKey]
	if !ok {
		return "", ErrNoAvatarURL
	}

	avatarURL, ok := v.(string)
	if !ok || avatarURL == "" {
		return "", ErrNoAvatarURL
	}

	return avatarURL, nil
}

func (GravatarAvatar) GetAvatarURL(c *client) (string, error) {
	v, ok := c.userData[emailKey]
	if !ok {
		return "", ErrNoAvatarURL
	}

	email, ok := v.(string)
	if !ok || email == "" {
		return "", ErrNoAvatarURL
	}

	m := md5.New()
	io.WriteString(m, strings.ToLower(email))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x", m.Sum(nil)), nil
}
