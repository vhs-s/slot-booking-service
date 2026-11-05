package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"slot-booking-service-api/internal/auth"
	authUseCase "slot-booking-service-api/internal/usecase/auth"
	"slot-booking-service-api/internal/usecase/auth/mocks"

	"go.uber.org/mock/gomock"
)

func TestCheckRefreshToken_ValidToken(t *testing.T) {
	const userID = "test-user-id"

	tokenPair, err := auth.GenerateTokenPair(userID)
	if err != nil {
		t.Fatalf("generate token pair: %v", err)
	}

	ctrl := gomock.NewController(t)
	tokenRepo := mocks.NewMockTokenRepo(ctrl)
	tokenRepo.EXPECT().
		GetCachedToken(gomock.Any(), userID, tokenPair.SessionID).
		Return(tokenPair.RefreshToken, nil).
		Times(1)

	manager := New(authUseCase.New(tokenRepo))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.AddCookie(&http.Cookie{
		Name:  "refresh_token",
		Value: tokenPair.RefreshToken,
	})

	gotUserID, gotSessionID, valid := manager.checkRefreshToken(request)

	if !valid {
		t.Fatal("expected matching refresh token to be valid")
	}
	if gotUserID != userID {
		t.Errorf("user ID = %q, want %q", gotUserID, userID)
	}
	if gotSessionID != tokenPair.SessionID {
		t.Errorf("session ID = %q, want %q", gotSessionID, tokenPair.SessionID)
	}
}
