package cliaas

type VMDeleter interface {
	Delete(vmIdentifier string) error
}

type VMReplacer interface {
	Replace(vmIdentifier string, imageIdentifier string) error
}
