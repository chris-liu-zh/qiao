package jwt

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/chris-liu-zh/qiao/Http"
	"github.com/chris-liu-zh/qiao/tools"
)

type Auth struct {
	issuer     string
	key        []byte
	accessExp  time.Duration
	refreshExp time.Duration
	CtxKey     Http.CtxKey
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

const (
	issuer                  = "app"
	accessExp               = 10 * time.Minute
	refreshExp              = 7 * 24 * time.Hour
	defaultKey              = "2D5JWUEGWWFK05JB74W1YGP9OF4L205F"
	tokenCtxKey Http.CtxKey = "token_info"
)

type AuthConfig struct {
	Issuer     string
	AccessExp  time.Duration
	RefreshExp time.Duration
	CtxKey     Http.CtxKey
	key        string
}

type Options func(*AuthConfig)

func WithIssuer(issuer string) Options {
	return func(c *AuthConfig) {
		c.Issuer = issuer
	}
}

func WithAccessExp(exp time.Duration) Options {
	return func(c *AuthConfig) {
		c.AccessExp = exp
	}
}

func WithRefreshExp(exp time.Duration) Options {
	return func(c *AuthConfig) {
		c.RefreshExp = exp
	}
}

func WithCtxKey(tokenCtxKey Http.CtxKey) Options {
	return func(c *AuthConfig) {
		c.CtxKey = Http.CtxKey(tokenCtxKey)
	}
}

func WithKey(key string) Options {
	return func(c *AuthConfig) {
		c.key = key
	}
}

func NewIssuer(opts ...Options) {
	ac := &AuthConfig{
		Issuer:     issuer,
		AccessExp:  accessExp,
		RefreshExp: refreshExp,
		CtxKey:     tokenCtxKey,
		key:        defaultKey,
	}
	for _, opt := range opts {
		opt(ac)
	}
	SetAuth(ac.Issuer, ac.AccessExp, ac.RefreshExp, ac.key)
}

func SetAuth(issuer string, accessExp, refreshExp time.Duration, key string) {
	if _, ok := authList[issuer]; ok {
		slog.Warn("issuer exist", "issuer", issuer)
	}
	authList[issuer] = &Auth{
		issuer:     issuer,
		key:        []byte(key),
		accessExp:  accessExp,
		refreshExp: refreshExp,
		CtxKey:     Http.CtxKey(issuer),
	}
}

func GetCtxKey(issuer string) (Http.CtxKey, bool) {
	auth, ok := authList[issuer]
	if !ok {
		return "", false
	}
	return auth.CtxKey, ok
}

func GetAccessExp(issuer string) (time.Duration, bool) {
	auth, ok := authList[issuer]
	if !ok {
		return 0, false
	}
	return auth.accessExp, true
}

func GetRefreshExp(issuer string) (time.Duration, bool) {
	auth, ok := authList[issuer]
	if !ok {
		return 0, false
	}
	return auth.refreshExp, true
}

// DefaultSign /**
func DefaultSign(sign, appKey, secret string, ts time.Time, timeDiff time.Duration) error {
	now := time.Now()
	// 检查时间戳是否在有效时间范围内
	if ts.Before(now.Add(-timeDiff)) || ts.After(now.Add(timeDiff)) {
		return errors.New("beyond the valid time range")
	}
	s := fmt.Sprintf("%s%s%d", appKey, secret, ts.Unix())
	localSign := strings.ToUpper(tools.MD5(s))

	if localSign != sign {
		return errors.New("sign error")
	}
	return nil
}

// CreateToken 创建新的 DefaultToken
func CreateToken(issuer string, claimsOption ...ClaimsOption) (t DefaultToken, err error) {
	if auth, ok := authList[issuer]; ok {
		uuid := tools.UUIDV7()
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
