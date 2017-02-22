package main

func main() {
}

type IaasClient interface {
	GetInstanceID() (string, error)
}
