package cliaas

type VMDeleter interface {
	Delete(identifier string) error
}
