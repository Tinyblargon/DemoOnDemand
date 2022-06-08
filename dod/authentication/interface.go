package authentication

var Main Backend

type Backend interface {
	Authenticate(userName, password string) (group string, succecfulLogin bool, err error)
}
