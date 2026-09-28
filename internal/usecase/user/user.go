package user

type UserUseCase struct {
	UserRepository  UserRepo
	TokenRepository TokenRepo
}

func New(userRepo UserRepo, tokenRepo TokenRepo) *UserUseCase {
	return &UserUseCase{
		UserRepository:  userRepo,
		TokenRepository: tokenRepo,
	}
}
