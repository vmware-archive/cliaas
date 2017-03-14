package cliaas

type VMReplacer interface {
	Replace(identifier string) error
}

type VMDeleter interface {
	Delete(identifier string) error
}
