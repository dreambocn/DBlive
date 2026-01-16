package service

import (
    "context"
    "errors"
    "time"

    "dblive/internal/config"
    "dblive/internal/model"
    "dblive/internal/repo"
    "dblive/internal/util"
)

type AuthService struct {
    users  *repo.UserRepo
    tokens *repo.TokenRepo
    cfg    config.Config
}

type TokenPair struct {
    AccessToken  string `json:"access_token"`
    RefreshToken string `json:"refresh_token"`
    ExpiresIn    int64  `json:"expires_in"`
}

func NewAuthService(users *repo.UserRepo, tokens *repo.TokenRepo, cfg config.Config) *AuthService {
    return &AuthService{users: users, tokens: tokens, cfg: cfg}
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*TokenPair, error) {
    user, err := s.users.GetByUsername(username)
    if err != nil || user == nil {
        return nil, errors.New("invalid credentials")
    }

    if !util.CheckPassword(password, user.PasswordHash) {
        return nil, errors.New("invalid credentials")
    }

    return s.issueTokens(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
    tokenHash := util.HashToken(refreshToken)
    now := time.Now().UTC()

    stored, err := s.tokens.GetValid(tokenHash, now)
    if err != nil || stored == nil {
        return nil, errors.New("invalid refresh token")
    }

    user, err := s.users.GetByID(stored.UserID)
    if err != nil || user == nil {
        return nil, errors.New("invalid refresh token")
    }

    if err := s.tokens.Revoke(tokenHash, now); err != nil {
        return nil, errors.New("refresh failed")
    }

    return s.issueTokens(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
    tokenHash := util.HashToken(refreshToken)
    return s.tokens.Revoke(tokenHash, time.Now().UTC())
}

func (s *AuthService) UserByID(ctx context.Context, id int64) (*model.User, error) {
    return s.users.GetByID(id)
}

func (s *AuthService) issueTokens(ctx context.Context, user *model.User) (*TokenPair, error) {
    accessToken, err := util.GenerateAccessToken(user, s.cfg.JWTSecret, s.cfg.AccessTTL)
    if err != nil {
        return nil, errors.New("token generation failed")
    }

    refreshToken, err := util.GenerateRefreshToken()
    if err != nil {
        return nil, errors.New("token generation failed")
    }

    now := time.Now().UTC()
    tokenHash := util.HashToken(refreshToken)
    record := &model.RefreshToken{
        UserID:    user.ID,
        TokenHash: tokenHash,
        ExpiresAt: now.Add(s.cfg.RefreshTTL),
        CreatedAt: now,
    }

    if err := s.tokens.Create(record); err != nil {
        return nil, errors.New("token store failed")
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
        ExpiresIn:    int64(s.cfg.AccessTTL.Seconds()),
    }, nil
}
