package login

type Command struct {
	Email     string
	Password  string
	UserAgent *string
	IP        *string
}
