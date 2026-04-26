package service

type AccountNumberGenerator interface {
	Next() (string, error)
}
