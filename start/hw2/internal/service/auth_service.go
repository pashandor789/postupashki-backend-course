package service

import (
	"errors"
	"hw2/internal/models"
	"hw2/internal/repository"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret []byte
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{userRepo: userRepo, jwtSecret: []byte(jwtSecret)}
}

func (s *AuthService) Register(username, password string) (*models.User, error) {
	if existing := s.userRepo.FindUser(username); existing != nil {
		return nil, errors.New("username already taken")
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:       username,
		HashedPassword: string(hashedBytes),
	}

	s.userRepo.SaveUser(user)

	return user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user := s.userRepo.FindUser(username)

	if user == nil {
		return "", errors.New("invalid username or password")
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))

	if err != nil {
		return "", errors.New("invalid username or password")
	}

	token, err := s.generateJWT(username)
	if err != nil {
		return "", err
	}

	return token, nil

}

func (s *AuthService) generateJWT(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(s.jwtSecret)
}
