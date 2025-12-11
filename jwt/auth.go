package jwt

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/chris-liu-zh/qiao"
)

type Auth struct {
	issuer     string
	key        []byte
	accessExp  time.Duration
	refreshExp time.Duration
}

type DefaultToken struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken,omitempty"`
}

var (
	authList          = make(map[string]*Auth)         // 认证列表
	ErrIssuerNotExist = errors.New("issuer not exist") // 发行者不存在
	ErrIssuerExist    = errors.New("issuer exist")     // 发行者已存在
)

type ClaimsOption func(*DefaultClaims)

type DefaultClaims struct {
	RegisteredClaims
	UserInfo any `json:"user_info,omitempty"`
}

func WithSubject(sub string) ClaimsOption {
	return func(c *DefaultClaims) {
		c.Subject = sub
	}
}

func WithUserInfo(userInfo any) ClaimsOption {
	return func(c *DefaultClaims) {
		c.UserInfo = userInfo
	}
}

func getNumericDate(exp time.Duration) *NumericDate {
	if exp > 0 {
		return NewNumericDate(time.Now().Add(exp))
	}
	return NewNumericDate(time.Now())
}

func SetAuth(issuer string, accessExp, refreshExp time.Duration, key string) error {
	if _, ok := authList[issuer]; ok {
		slog.Warn("issuer exist", "issuer", issuer)
	}
	authList[issuer] = &Auth{
		issuer:     issuer,
		key:        []byte(key),
		accessExp:  accessExp,
		refreshExp: refreshExp,
	}
	return nil
}

// DefaultSign /**
func DefaultSign(sign, appKey, secret string, ts time.Time, timeDiff time.Duration) error {
	now := time.Now()
	// 检查时间戳是否在有效时间范围内
	if ts.Before(now.Add(-timeDiff)) || ts.After(now.Add(timeDiff)) {
		return errors.New("beyond the valid time range")
	}
	s := fmt.Sprintf("%s%s%d", appKey, secret, ts.Unix())
	localSign := strings.ToUpper(qiao.MD5(s))

	if localSign != sign {
		return errors.New("sign error")
	}
	return nil
}

// CreateToken 创建新的 DefaultToken
func CreateToken(issuer string, claimsOption ...ClaimsOption) (t DefaultToken, err error) {
	if auth, ok := authList[issuer]; ok {
		uuid := qiao.UUIDV7()
		ac := DefaultClaims{
			RegisteredClaims: RegisteredClaims{
				ExpiresAt: getNumericDate(auth.accessExp),
				Issuer:    auth.issuer,
				ID:        uuid.String(),
			},
		}
		for _, option := range claimsOption {
			option(&ac)
		}
		t.Token, err = NewToken(SignMethodHS256, ac).Sign(auth.key)
		if err != nil {
			return
		}
		if auth.refreshExp > 0 {
			ac.ExpiresAt = getNumericDate(auth.refreshExp)
			t.RefreshToken, err = NewToken(SignMethodHS256, ac).Sign(auth.key)
			if err != nil {
				return
			}
		}
		return t, nil
	}
	return t, ErrIssuerNotExist
}

// CheckToken 验证Token
func CheckToken(issuer, token string, userinfo any) error {
	if auth, ok := authList[issuer]; ok {
		c := &DefaultClaims{
			UserInfo: userinfo,
		}
		if err := VerifyToken(token, c, auth.key); err != nil {
			return err
		}
		// if GetInvalidToken(c.ID) {
		// 	return ErrTokenRevoked
		// }
		if c.Issuer != auth.issuer {
			return ErrTokenInvalidIssuer
		}
		return nil
	}
	return ErrIssuerNotExist
}

func GetClaims(issuer, token string) (claims *DefaultClaims, err error) {
	if auth, ok := authList[issuer]; ok {
		claims = &DefaultClaims{}
		if err = VerifyToken(token, claims, auth.key); err != nil {
			return
		}
		return
	}
	return nil, ErrIssuerNotExist
}

// RefreshToken 刷新Token
func RefreshToken(issuer, accessToken, refreshToken string) (t DefaultToken, err error) {
	accessClaims, err := GetClaims(issuer, accessToken)
	if err != nil {
		if errors.Is(err, ErrTokenExpired) {
			refreshClaims, err := GetClaims(issuer, refreshToken)
			if err != nil {
				return t, err
			}
			if refreshClaims.ID != accessClaims.ID && refreshClaims.Issuer != accessClaims.Issuer {
				return t, ErrTokenNotMatch
			}
			auth := authList[issuer]
			accessClaims.ExpiresAt = getNumericDate(auth.accessExp)
			refreshClaims.ExpiresAt = getNumericDate(auth.refreshExp)
			if t.Token, err = NewToken(SignMethodHS256, accessClaims).Sign(auth.key); err != nil {
				return t, err
			}
			if t.RefreshToken, err = NewToken(SignMethodHS256, refreshClaims).Sign(auth.key); err != nil {
				return t, err
			}
			return t, nil
		}
		return t, err
	}
	return DefaultToken{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}
