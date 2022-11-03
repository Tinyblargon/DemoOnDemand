package ldap

import (
	"crypto/tls"
	"fmt"

	ldap "github.com/go-ldap/ldap/v3"
)

type Settings struct {
	URL                string
	BindDN             string
	BindCredential     string
	InsecureSkipVerify bool
	UsernameAttribute  string
	UserGroup          *Settings_Group
	AdminGroup         *Settings_Group
}

type Settings_Group struct {
	UsersDN string
}

func New(settings *Settings) (*Settings, error) {
	l, err := settings.NewSession()
	if err != nil {
		return nil, err
	}
	l.Close()
	return settings, nil
}

func (s *Settings) Authenticate(userName, password string) (group string, successfulLogin bool, err error) {
	newSettings := Settings{
		URL:                s.URL,
		InsecureSkipVerify: s.InsecureSkipVerify,
	}
	l, err := s.NewSession()
	if err != nil {
		return
	}
	result, err := findUser(l, s.AdminGroup.UsersDN, userName, s.UsernameAttribute)
	if err != nil {
		l.Close()
		return
	}
	if len(result.Entries) != 0 {
		l.Close()
		return newSettings.authenticateSubtask(result, password, "admin")
	}
	result, err = findUser(l, s.UserGroup.UsersDN, userName, s.UsernameAttribute)
	if err != nil {
		l.Close()
		return
	}
	l.Close()
	if len(result.Entries) != 0 {
		return newSettings.authenticateSubtask(result, password, "user")
	}
	return
}

func (s *Settings) authenticateSubtask(result *ldap.SearchResult, password, group string) (string, bool, error) {
	s.BindDN = result.Entries[0].DN
	s.BindCredential = password
	connection, err := s.NewSession()
	connection.Close()
	if err != nil {
		return "", false, nil
	}
	return group, true, err
}

func findUser(l *ldap.Conn, baseDN, user, usernameAttribute string) (result *ldap.SearchResult, err error) {
	filter := fmt.Sprintf("(%s=%s)", usernameAttribute, ldap.EscapeFilter(user))
	result, err = l.Search(ldap.NewSearchRequest(baseDN, ldap.ScopeWholeSubtree, 0, 0, 0, false, filter, []string{usernameAttribute}, []ldap.Control{}))
	return
}

func (s *Settings) NewSession() (l *ldap.Conn, err error) {
	if s.URL[0:8] == "ldaps://" && s.InsecureSkipVerify {
		l, err = ldap.DialURL(s.URL, ldap.DialWithTLSConfig(&tls.Config{InsecureSkipVerify: true}))
	} else {
		l, err = ldap.DialURL(s.URL)
	}
	if err != nil {
		return
	}
	err = l.Bind(s.BindDN, s.BindCredential)
	return
}
