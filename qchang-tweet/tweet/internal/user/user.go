package user

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail        = errors.New("invalid email")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrRequiredFirstname   = errors.New("firstname is required")
	ErrRequiredLastname    = errors.New("lastname is required")
	ErrRequiredDisplayname = errors.New("displayname is required")
)

type User struct {
	ID             uuid.UUID
	Email          string
	Firstname      string
	Lastname       string
	Displayname    string
	HashedPassword []byte
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}

func NewUser(email string, password string, firstname string, lastname string, displayname string) (*User, error) {
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}

	if firstname == "" {
		return nil, ErrRequiredFirstname
	}

	if lastname == "" {
		return nil, ErrRequiredLastname
	}

	if displayname == "" {
		return nil, ErrRequiredDisplayname
	}

	user := &User{
		ID:          uuid.New(),
		Email:       email,
		Firstname:   firstname,
		Lastname:    lastname,
		Displayname: displayname,
		CreatedAt:   time.Now(),
		UpdatedAt:   nil,
	}
	err := user.setPassword(password)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) setPassword(password string) error {
	if len(password) < 8 {
		return ErrInvalidPassword
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.HashedPassword = hashedPassword
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword(u.HashedPassword, []byte(password))
	return err == nil
}

func isValidEmail(email string) bool {
	// This is a simple email validation regex, it may not cover all cases
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
