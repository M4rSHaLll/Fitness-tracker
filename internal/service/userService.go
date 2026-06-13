package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/repository"
	"Fitness-tracker/internal/validation"
	"errors"
	"fmt"
	"strings"
	"time"
)

type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(telegramID int64, username string) (*model.User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, fmt.Errorf("%w: username is required", ErrInvalidInput)
	}

	user := model.NewUser(
		telegramID,
		username,
	)

	err := s.repo.Create(user)
	if err != nil {
		if errors.Is(err, repository.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, fmt.Errorf("can't create new user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("can't get user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateProfile(userID int64, weight, height float64, age int64) error {
	user, err := s.repo.GetByID(userID)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("can't get user: %w", err)
	}

	if err := validation.ValidationWeight(weight); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if err := validation.ValidationHeight(height); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}
	if err := validation.ValidationAge(age); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	user.Age = age
	user.Height = height
	user.Weight = weight
	user.UpdatedAt = time.Now()

	err = s.repo.Update(user)

	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("can't update user profile: %w", err)
	}

	return nil
}
