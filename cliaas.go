package cliaas

type VMDeleter interface {
	Delete(identifier string) error
}

type VMReplacer interface {
	Replace(identifier string) error
}
