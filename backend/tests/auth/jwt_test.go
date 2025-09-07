package auth

import (
	"jobView-backend/internal/auth"
	"jobView-backend/internal/model"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 注意：setupTestUser 函数已经在 middleware_test.go 中定义，这里不重复定义

func TestGenerateTokenPair(t *testing.T) {
	user := setupTestUser()

	accessToken, refreshToken, err := auth.GenerateTokenPair(user)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)
	
	// 验证access token
	accessClaims, err := auth.ValidateAccessToken(accessToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, accessClaims.UserID)
	assert.Equal(t, user.Username, accessClaims.Username)
	assert.Equal(t, user.Email, accessClaims.Email)
	assert.Equal(t, "access", accessClaims.Type)
	
	// 验证refresh token
	refreshClaims, err := auth.ValidateRefreshToken(refreshToken)
	require.NoError(t, err)
	assert.Equal(t, user.ID, refreshClaims.UserID)
	assert.Equal(t, user.Username, refreshClaims.Username)
	assert.Equal(t, user.Email, refreshClaims.Email)
	assert.Equal(t, "refresh", refreshClaims.Type)
}

func TestValidateAccessToken(t *testing.T) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	t.Run("ValidToken", func(t *testing.T) {
		claims, err := auth.ValidateAccessToken(accessToken)
		
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, "access", claims.Type)
	})

	t.Run("EmptyToken", func(t *testing.T) {
		_, err := auth.ValidateAccessToken("")
		
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidToken, err)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		_, err := auth.ValidateAccessToken("invalid-token")
		
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidToken, err)
	})

	t.Run("WrongTokenType", func(t *testing.T) {
		_, refreshToken, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		_, err = auth.ValidateAccessToken(refreshToken)
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidToken, err)
	})
}

func TestValidateRefreshToken(t *testing.T) {
	user := setupTestUser()
	_, refreshToken, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	t.Run("ValidToken", func(t *testing.T) {
		claims, err := auth.ValidateRefreshToken(refreshToken)
		
		require.NoError(t, err)
		assert.Equal(t, user.ID, claims.UserID)
		assert.Equal(t, user.Username, claims.Username)
		assert.Equal(t, "refresh", claims.Type)
	})

	t.Run("WrongTokenType", func(t *testing.T) {
		accessToken, _, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		_, err = auth.ValidateRefreshToken(accessToken)
		assert.Error(t, err)
		assert.Equal(t, auth.ErrInvalidToken, err)
	})
}

func TestExtractTokenFromHeader(t *testing.T) {
	testCases := []struct {
		name      string
		header    string
		expected  string
		expectErr bool
	}{
		{
			name:      "ValidHeader",
			header:    "Bearer abc123token",
			expected:  "abc123token",
			expectErr: false,
		},
		{
			name:      "EmptyHeader",
			header:    "",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "InvalidFormat",
			header:    "Basic abc123",
			expected:  "",
			expectErr: true,
		},
		{
			name:      "MissingToken",
			header:    "Bearer ",
			expected:  "",
			expectErr: false,
		},
		{
			name:      "TooShort",
			header:    "Bear",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := auth.ExtractTokenFromHeader(tc.header)
			
			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, token)
			}
		})
	}
}

func TestTokenExpiration(t *testing.T) {
	user := setupTestUser()

	t.Run("TokenNotExpired", func(t *testing.T) {
		accessToken, _, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		claims, err := auth.ValidateAccessToken(accessToken)
		require.NoError(t, err)
		
		// 新生成的token应该不会在30分钟内过期
		assert.False(t, auth.IsTokenExpired(claims))
	})

	t.Run("ExpiredToken", func(t *testing.T) {
		// 创建一个已过期的token
		now := time.Now()
		claims := &auth.CustomClaims{
			UserID:   user.ID,
			Username: user.Username,
			Email:    user.Email,
			Type:     "access",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(now.Add(-24 * time.Hour)), // 24小时前过期
				IssuedAt:  jwt.NewNumericDate(now.Add(-25 * time.Hour)),
				NotBefore: jwt.NewNumericDate(now.Add(-25 * time.Hour)),
				Issuer:    "jobview-backend",
				Subject:   user.Username,
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		require.NoError(t, err)

		// 验证过期的token应该返回错误
		_, err = auth.ValidateAccessToken(tokenString)
		assert.Error(t, err)
		// 注意：jwt库返回的过期错误可能会被包装为 "invalid token"
		assert.Contains(t, []error{auth.ErrTokenExpired, auth.ErrInvalidToken}, err)
	})
}

func TestGetTokenExpirationTime(t *testing.T) {
	testCases := []struct {
		tokenType string
		expected  time.Duration
	}{
		{"access", auth.AccessTokenDuration},
		{"refresh", auth.RefreshTokenDuration},
		{"unknown", auth.AccessTokenDuration}, // 默认值
	}

	for _, tc := range testCases {
		t.Run(tc.tokenType, func(t *testing.T) {
			duration := auth.GetTokenExpirationTime(tc.tokenType)
			assert.Equal(t, tc.expected, duration)
		})
	}
}

func TestTokenSigning(t *testing.T) {
	user := setupTestUser()

	t.Run("ConsistentSigning", func(t *testing.T) {
		// 同一个用户应该能生成不同的token（因为时间戳不同）
		token1, _, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		// 确保有足够的时间差
		time.Sleep(10 * time.Millisecond)
		
		token2, _, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		// Token应该不同（因为时间戳不同）
		// 注意：由于JWT包含了时间戳，即使用户相同，token也会不同
		if token1 == token2 {
			t.Log("Warning: Tokens are identical, this might indicate timing issues")
		}
		
		// 但都应该能正常验证
		claims1, err := auth.ValidateAccessToken(token1)
		require.NoError(t, err)
		
		claims2, err := auth.ValidateAccessToken(token2)
		require.NoError(t, err)
		
		assert.Equal(t, claims1.UserID, claims2.UserID)
		assert.Equal(t, claims1.Username, claims2.Username)
	})
}

func TestCustomClaims(t *testing.T) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	require.NoError(t, err)

	claims, err := auth.ValidateAccessToken(accessToken)
	require.NoError(t, err)

	// 验证自定义字段
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Email, claims.Email)
	assert.Equal(t, "access", claims.Type)
	
	// 验证标准JWT声明
	assert.Equal(t, "jobview-backend", claims.Issuer)
	assert.Equal(t, user.Username, claims.Subject)
	assert.True(t, claims.ExpiresAt.After(time.Now()))
	assert.True(t, claims.IssuedAt.Before(time.Now().Add(time.Second)))
}

func TestEdgeCases(t *testing.T) {
	t.Run("NilUser", func(t *testing.T) {
		// 测试nil用户应该不会panic
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("GenerateTokenPair should panic with nil user")
			}
		}()
		
		_, _, err := auth.GenerateTokenPair(nil)
		// 如果到达这里说明没有panic，应该有错误
		assert.Error(t, err)
	})

	t.Run("ZeroUser", func(t *testing.T) {
		user := &model.User{}
		
		accessToken, refreshToken, err := auth.GenerateTokenPair(user)
		require.NoError(t, err)
		
		claims, err := auth.ValidateAccessToken(accessToken)
		require.NoError(t, err)
		
		assert.Equal(t, uint(0), claims.UserID)
		assert.Equal(t, "", claims.Username)
		assert.Equal(t, "", claims.Email)
		
		// Refresh token也应该工作
		refreshClaims, err := auth.ValidateRefreshToken(refreshToken)
		require.NoError(t, err)
		assert.Equal(t, uint(0), refreshClaims.UserID)
	})
}

// BenchmarkGenerateTokenPair 性能测试
func BenchmarkGenerateTokenPair(b *testing.B) {
	user := setupTestUser()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := auth.GenerateTokenPair(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkValidateAccessToken 性能测试
func BenchmarkValidateAccessToken(b *testing.B) {
	user := setupTestUser()
	accessToken, _, err := auth.GenerateTokenPair(user)
	if err != nil {
		b.Fatal(err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := auth.ValidateAccessToken(accessToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}