package handler

import (
	"encoding/json"
	"net/http"

	"Go-Next-WebRTC/internal/adapter/http/dto"
	"Go-Next-WebRTC/internal/adapter/http/middleware"
	"Go-Next-WebRTC/internal/application/usecase"
	"Go-Next-WebRTC/internal/domain/entity"
)

type AuthHandler struct {
	authUseCase usecase.AuthUseCase
}

func NewAuthHandler(authUseCase usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		authUseCase: authUseCase,
	}
}

// Register 新規登録
// @Summary ユーザー登録
// @Description 新規ユーザーを登録し、認証トークンを返却
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "登録情報"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	// バリデーション
	if err := h.validateRegisterRequest(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Validation failed", err)
		return
	}

	tokens, err := h.authUseCase.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		switch err {
		case entity.ErrEmailAlreadyExists:
			h.respondWithError(w, http.StatusConflict, "Email already exists", nil)
		default:
			h.respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		}
		return
	}

	h.respondWithJSON(w, http.StatusCreated, dto.ToAuthResponse(tokens))
}

// Login ログイン
// @Summary ユーザーログイン
// @Description メールアドレスとパスワードでログイン
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "ログイン情報"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	tokens, err := h.authUseCase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		switch err {
		case entity.ErrInvalidCredentials:
			h.respondWithError(w, http.StatusUnauthorized, "Invalid email or password", nil)
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Login failed", nil)
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.ToAuthResponse(tokens))
}

// RefreshToken リフレッシュトークン
// @Summary トークン更新
// @Description リフレッシュトークンを使用してアクセストークンを更新
// @Tags auth
// @Accept json
// @Produce json
// @Param X-Refresh-Token header string false "リフレッシュトークン"
// @Param request body dto.RefreshTokenRequest false "リフレッシュトークン（ボディ）"
// @Success 200 {object} dto.AuthResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// ヘッダーから取得を優先
	refreshToken := r.Header.Get("X-Refresh-Token")
	
	// ヘッダーにない場合はボディから取得
	if refreshToken == "" {
		var req dto.RefreshTokenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Refresh token required", nil)
			return
		}
		refreshToken = req.RefreshToken
	}

	if refreshToken == "" {
		h.respondWithError(w, http.StatusBadRequest, "Refresh token required", nil)
		return
	}

	tokens, err := h.authUseCase.RefreshToken(r.Context(), refreshToken)
	if err != nil {
		switch err {
		case entity.ErrInvalidToken, entity.ErrTokenExpired:
			h.respondWithError(w, http.StatusUnauthorized, "Invalid or expired refresh token", nil)
		default:
			h.respondWithError(w, http.StatusInternalServerError, "Token refresh failed", nil)
		}
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.ToAuthResponse(tokens))
}

// Logout ログアウト
// @Summary ログアウト
// @Description 現在のセッションを終了
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.authUseCase.Logout(r.Context(), userID); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Logout failed", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Successfully logged out",
	})
}

// GetCurrentUser 現在のユーザー情報取得
// @Summary 現在のユーザー情報取得
// @Description 認証済みユーザーの情報を取得
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} dto.UserResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	user, err := h.authUseCase.GetUserByID(r.Context(), userID)
	if err != nil {
		h.respondWithError(w, http.StatusNotFound, "User not found", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.ToUserResponse(user))
}

// UpdateProfile プロフィール更新
// @Summary プロフィール更新
// @Description ユーザーのプロフィール情報を更新
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.UpdateProfileRequest true "プロフィール情報"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/auth/profile [put]
func (h *AuthHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req dto.UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	user, err := h.authUseCase.UpdateUserProfile(r.Context(), userID, req.Name)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.ToUserResponse(user))
}

// ChangePassword パスワード変更
// @Summary パスワード変更
// @Description 現在のパスワードを確認して新しいパスワードに変更
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body dto.ChangePasswordRequest true "パスワード変更情報"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/auth/password [put]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	var req dto.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.authUseCase.ChangePassword(r.Context(), userID, req.OldPassword, req.NewPassword); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Password successfully changed",
	})
}

// RequestPasswordReset パスワードリセット要求
// @Summary パスワードリセット要求
// @Description パスワードリセット用のトークンを生成してメール送信
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.PasswordResetRequest true "メールアドレス"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/auth/password-reset/request [post]
func (h *AuthHandler) RequestPasswordReset(w http.ResponseWriter, r *http.Request) {
	var req dto.PasswordResetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	_, err := h.authUseCase.RequestPasswordReset(r.Context(), req.Email)
	if err != nil {
		// セキュリティのため、エラーの詳細は返さない
	}

	// セキュリティのため、常に成功レスポンスを返す
	h.respondWithJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "If an account exists with this email, a password reset link has been sent",
	})
}

// ResetPassword パスワードリセット実行
// @Summary パスワードリセット実行
// @Description トークンを使用してパスワードをリセット
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.ResetPasswordRequest true "リセット情報"
// @Success 200 {object} dto.MessageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /api/auth/password-reset/confirm [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req dto.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", nil)
		return
	}

	if err := h.authUseCase.ResetPassword(r.Context(), req.Token, req.NewPassword); err != nil {
		h.respondWithError(w, http.StatusBadRequest, err.Error(), nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "Password successfully reset",
	})
}

// GetActiveSessions アクティブセッション一覧取得
// @Summary アクティブセッション一覧
// @Description 現在のユーザーのアクティブなセッション一覧を取得
// @Tags auth
// @Security BearerAuth
// @Success 200 {array} dto.SessionResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/sessions [get]
func (h *AuthHandler) GetActiveSessions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	sessions, err := h.authUseCase.GetActiveSessions(r.Context(), userID)
	if err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get sessions", nil)
		return
	}

	// TODO: セッションレスポンスの変換
	h.respondWithJSON(w, http.StatusOK, sessions)
}

// RevokeAllSessions 全セッション無効化
// @Summary 全セッション無効化
// @Description 現在のセッションを除く全てのセッションを無効化
// @Tags auth
// @Security BearerAuth
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/auth/sessions/revoke-all [post]
func (h *AuthHandler) RevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	userID := h.getUserIDFromContext(r)
	if userID == 0 {
		h.respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	if err := h.authUseCase.RevokeAllSessions(r.Context(), userID); err != nil {
		h.respondWithError(w, http.StatusInternalServerError, "Failed to revoke sessions", nil)
		return
	}

	h.respondWithJSON(w, http.StatusOK, dto.MessageResponse{
		Message: "All sessions have been revoked",
	})
}

// ヘルパーメソッド

func (h *AuthHandler) getUserIDFromContext(r *http.Request) int64 {
	userID, ok := r.Context().Value("user_id").(int64)
	if !ok {
		// 新しいミドルウェアのキーも試す
		userID, ok = middleware.GetUserIDFromContext(r.Context())
		if !ok {
			return 0
		}
	}
	return userID
}

func (h *AuthHandler) respondWithError(w http.ResponseWriter, code int, message string, details map[string]string) {
	h.respondWithJSON(w, code, dto.ErrorResponse{
		Error:   message,
		Details: details,
	})
}

func (h *AuthHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) validateRegisterRequest(req *dto.RegisterRequest) map[string]string {
	errors := make(map[string]string)

	if req.Email == "" {
		errors["email"] = "Email is required"
	}
	if req.Password == "" {
		errors["password"] = "Password is required"
	} else if len(req.Password) < 8 {
		errors["password"] = "Password must be at least 8 characters"
	}
	if req.Name == "" {
		errors["name"] = "Name is required"
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}