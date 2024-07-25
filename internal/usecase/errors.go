package usecase

import "errors"

var (
	ErrAlreadyExists                 = errors.New("already exists")
	ErrInvalidClientRequest          = errors.New("invalid client request")
	ErrInvalidOpenIDConnectParameter = errors.New("invalid openID connect parameter")
	ErrAffiliatorNotFound            = errors.New("affiliator not found")
	ErrUnAuthorized                  = errors.New("unauthorized")
	ErrNotFound                      = errors.New("not found")
)
