package system

import (
	"golang.org/x/crypto/bcrypt"
)

func ListUsers() ([]string, error) {
	s, err := loadUsers()
	if err != nil {
		return nil, err
	}

	out := make([]string, 0, len(s.Users))
	for u := range s.Users {
		out = append(out, u)
	}
	return out, nil
}

func DeleteUser(username string) error {
	s, err := loadUsers()
	if err != nil {
		return err
	}

	if _, ok := s.Users[username]; !ok {
		return ErrUserNotFound
	}

	delete(s.Users, username)
	return saveUsers(s)
}

func AddUser(username, password string) error {
	s, err := loadUsers()
	if err != nil {
		return err
	}

	if _, ok := s.Users[username]; ok {
		return ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	s.Users[username] = string(hash)
	return saveUsers(s)
}

func Authenticate(username, password string) error {
	s, err := loadUsers()
	if err != nil {
		return err
	}

	hash, ok := s.Users[username]
	if !ok {
		return ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		return ErrBadCredential
	}

	return nil
}
