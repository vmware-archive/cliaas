package cliaas

type Config struct {
	AWS struct {
		AccessKeyID     string `yaml:"access_key_id"`
		SecretAccessKey string `yaml:"secret_access_key"`
		Region          string `yaml:"region"`
		VPC             string `yaml:"vpc"`
		AMI             string `yaml:"ami"`
	} `yaml:"aws"`
}
