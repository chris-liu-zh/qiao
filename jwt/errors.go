package jwt

import "errors"

var (
	ErrHashUnavailable           = errors.New("the requested hash function is unavailable") //请求的哈希函数不可用
	ErrTokenMalformed            = errors.New("token is malformed")                         //令牌格式错误
	ErrTokenUnverifiable         = errors.New("token is unverifiable")                      //令牌无法验证
	ErrTokenSignatureInvalid     = errors.New("token signature is invalid")                 //令牌签名无效
	ErrTokenRequiredClaimMissing = errors.New("token is missing required claim")            //令牌缺失必填声明
	ErrTokenInvalidAudience      = errors.New("token has invalid audience")                 //令牌受众无效
	ErrTokenExpired              = errors.New("token is expired")                           //令牌已过期
	ErrTokenUsedBeforeIssued     = errors.New("token used before issued")                   //令牌已使用
	ErrTokenInvalidIssuer        = errors.New("token has invalid issuer")                   //令牌发行者无效
	ErrTokenInvalidSubject       = errors.New("token has invalid subject")                  //令牌主题无效
	ErrTokenNotValidYet          = errors.New("token is not valid yet")                     //令牌未生效
	ErrTokenInvalidId            = errors.New("token has invalid id")                       //令牌ID无效
	ErrTokenNotMatch             = errors.New("access token does not match refresh token")  //访问令牌与刷新令牌不匹配
	ErrTokenInvalidHeader        = errors.New("token has invalid header")                   //令牌头信息无效
	ErrTokenInvalidClaims        = errors.New("token has invalid claims")                   //令牌声明信息无效
	ErrTokenRevoked              = errors.New("token is revoked")                           //令牌已撤销
)
