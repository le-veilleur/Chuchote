package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/maxime/chuchote/application/dto"
	domainerrors "github.com/maxime/chuchote/domain/errors"
	"github.com/maxime/chuchote/domain/model"
	"github.com/maxime/chuchote/port/outbound"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     outbound.UserRepository
	jwtSecret []byte
}

func NewAuthService(users outbound.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret)}
}

func (s *AuthService) Register(ctx context.Context, cmd dto.RegisterCommand) (dto.TokenView, error) {
	if cmd.Username == "" || cmd.Password == "" {
		return dto.TokenView{}, domainerrors.ErrInvalidInput
	}

	_, err := s.users.FindByUsername(ctx, cmd.Username)
	if err == nil {
		return dto.TokenView{}, domainerrors.ErrUsernameExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(cmd.Password), bcrypt.DefaultCost)
	if err != nil {
		return dto.TokenView{}, err
	}

	user := model.User{
		ID:           model.UserID(uuid.NewString()),
		Username:     cmd.Username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.users.Save(ctx, user); err != nil {
		return dto.TokenView{}, err
	}

	return s.generateToken(user)
}

func (s *AuthService) Login(ctx context.Context, cmd dto.LoginCommand) (dto.TokenView, error) {
	user, err := s.users.FindByUsername(ctx, cmd.Username)
	if err != nil {
		return dto.TokenView{}, domainerrors.ErrUnauthorized
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(cmd.Password)); err != nil {
		return dto.TokenView{}, domainerrors.ErrUnauthorized
	}

	return s.generateToken(user)
}

func (s *AuthService) ValidateToken(_ context.Context, tokenStr string) (dto.UserClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, domainerrors.ErrInvalidToken
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return dto.UserClaims{}, domainerrors.ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return dto.UserClaims{}, domainerrors.ErrInvalidToken
	}

	return dto.UserClaims{
		UserID:   model.UserID(claims["sub"].(string)),
		Username: claims["username"].(string),
	}, nil
}

func (s *AuthService) generateToken(user model.User) (dto.TokenView, error) {
	claims := jwt.MapClaims{
		"sub":      string(user.ID),
		"username": user.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return dto.TokenView{}, err
	}
	return dto.TokenView{Token: signed, UserID: user.ID, Username: user.Username}, nil
}
