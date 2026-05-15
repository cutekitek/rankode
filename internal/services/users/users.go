package users

import (
	"context"
	"errors"

	apierror "rankode/internal/errors"
	"rankode/internal/models"
	db "rankode/internal/repository"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt" // Import the bcrypt library
)

type UserService struct {
	// We use the Querier interface to interact with the database
	q db.Querier
}

// NewUserService creates a new instance of UserService.
func NewUserService(q db.Querier) *UserService {
	return &UserService{q: q}
}

func (s *UserService) Register(ctx context.Context, dto models.CreateUserDTO) (db.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		return db.User{}, err
	}

	params := db.NewUserParams{
		Username:     dto.Username,
		Email:        dto.Email,
		PasswordHash: string(hashedPassword),
		Elo:          0,
	}

	user, err := s.q.NewUser(ctx, params)
	if err != nil {
		return db.User{}, err
	}

	if dto.Roles != 0 {
		if err := s.q.UpdateUserRole(ctx, db.UpdateUserRoleParams{ID: user.ID, Roles: dto.Roles}); err != nil {
			return db.User{}, err
		}
		user.Roles = dto.Roles
	}

	return user, nil
}

func (s *UserService) Authenticate(ctx context.Context, dto models.AuthUserDTO) (db.User, error) {
	user, err := s.q.GetUsersByEmailOrUsername(ctx, dto.Identifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return db.User{}, apierror.WrapErrorApi(errors.New("invalid credentials"), 401)
		}
		return db.User{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(dto.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return db.User{}, apierror.WrapErrorApi(errors.New("invalid credentials"), 401)
		}
		return db.User{}, err
	}

	return user, nil
}

func (s *UserService) GetLeaderboard(ctx context.Context) ([]db.GetUsersLeaderboardRow, error) {
	return s.q.GetUsersLeaderboard(ctx, 100)
}
