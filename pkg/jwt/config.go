package jwt

type Config struct {
	Secret   string
	Issuer   string
	Audience string
}
