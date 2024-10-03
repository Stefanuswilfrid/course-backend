package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Repository interface {
	SaveOTP(ctx context.Context, email string, otp string) error
	GetOTP(ctx context.Context, email string) (string, error)
	DeleteOTP(ctx context.Context, email string) error
	SaveResetPasswordToken(ctx context.Context, email string, token string) error
	GetResetPasswordToken(ctx context.Context, email string) (string, error)
	DeleteResetPasswordToken(ctx context.Context, email string) error
}

type repository struct {
	otpStore           map[string]string
	resetPasswordStore map[string]string
	mutex              sync.RWMutex
}

func NewRepository() Repository {
	return &repository{
		otpStore:           make(map[string]string),
		resetPasswordStore: make(map[string]string),
	}
}

func (r *repository) SaveOTP(ctx context.Context, email string, otp string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.otpStore[email] = otp

	// Simulate expiry after 10 minutes
	go func() {
		time.Sleep(10 * time.Minute)
		r.DeleteOTP(ctx, email)
	}()

	return nil
}

func (r *repository) GetOTP(ctx context.Context, email string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	otp, exists := r.otpStore[email]
	if !exists {
		return "", errors.New("OTP not found")
	}

	return otp, nil
}

func (r *repository) DeleteOTP(ctx context.Context, email string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.otpStore, email)
	return nil
}

func (r *repository) SaveResetPasswordToken(ctx context.Context, email string, token string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	r.resetPasswordStore[email] = token

	// Simulate expiry after 10 minutes
	go func() {
		time.Sleep(10 * time.Minute)
		r.DeleteResetPasswordToken(ctx, email)
	}()

	return nil
}

func (r *repository) GetResetPasswordToken(ctx context.Context, email string) (string, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	token, exists := r.resetPasswordStore[email]
	if !exists {
		return "", errors.New("Reset password token not found")
	}

	return token, nil
}

func (r *repository) DeleteResetPasswordToken(ctx context.Context, email string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	delete(r.resetPasswordStore, email)
	return nil
}
