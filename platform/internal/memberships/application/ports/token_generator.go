package ports

type TokenGenerator interface {
	Generate() (string, error)
	Hash(raw string) string
}
