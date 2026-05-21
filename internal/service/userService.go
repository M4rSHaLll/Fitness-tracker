package service

import (
	"Fitness-tracker/internal/model"
	"Fitness-tracker/internal/validation"
	"fmt"

	"Fitness-tracker/internal/repository"
	"errors"
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
	//validate and use repo.Create
	if username == "" {
		return nil, errors.New("username is required")
	}

	user := model.NewUser(
		telegramID,
		username,
	)

	err := s.repo.Create(user)
	if err != nil {
		return nil, fmt.Errorf("can't create new user: %w", err)
	}

	return user, nil
}

func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	user, err := s.repo.GetByID(id)

	if err != nil {
		return nil, fmt.Errorf("can't get user: %w", err)
	}

	return user, nil
}

func (s *UserService) UpdateProfile(userID int64, weight, height float64, age int64) error {
	user, err := s.repo.GetByID(userID)

	if err != nil {
		return fmt.Errorf("can't get user: %w", err)
	}

	if err := validation.ValidationWeight(weight); err != nil {
		return err
	}
	if err := validation.ValidationHeight(height); err != nil {
		return err
	}
	if err := validation.ValidationAge(age); err != nil {
		return err
	}

	user.Age = age
	user.Height = height
	user.Weight = weight

	err = s.repo.Update(user)

	if err != nil {
		return fmt.Errorf("can't update user profile: %w", err)
	}

	return nil
}
