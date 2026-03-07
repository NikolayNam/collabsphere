package ports

type RandomTokenGenerator interface {
	Generate() (string, error)
	Hash(raw string) string
}
