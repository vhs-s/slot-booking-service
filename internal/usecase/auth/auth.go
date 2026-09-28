package auth

type AuthUseCase struct {
	TokenRepository TokenRepo
}

func New(tokenRepo TokenRepo) *AuthUseCase {
	return &AuthUseCase{
		TokenRepository: tokenRepo,
	}
}
