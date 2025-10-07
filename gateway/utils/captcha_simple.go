package utils

import (
	"context"

	base64Captcha "github.com/mojocn/base64Captcha"
)

// Captcha 验证码结构体
type Captcha struct {
	store base64Captcha.Store
}

// NewCaptcha 创建新的验证码实例
func NewCaptcha(store base64Captcha.Store) *Captcha {
	return &Captcha{
		store: store,
	}
}

// Generate 生成验证码
// 返回验证码ID、Base64编码的图片字符串和答案
func (c *Captcha) Generate() (id, b64s, answer string, err error) {
	// 创建数字验证码驱动
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)

	// 创建验证码
	captcha := base64Captcha.NewCaptcha(driver, c.store)

	// 生成验证码
	return captcha.Generate()
}

// Verify 验证用户输入的验证码
// id: 验证码ID
// answer: 用户输入的答案
func (c *Captcha) Verify(id, answer string) bool {
	// 验证验证码，验证后清除
	return c.store.Verify(id, answer, true)
}

// VerifyWithContext 支持context的验证码验证
func (c *Captcha) VerifyWithContext(ctx context.Context, id, answer string) bool {
	// 对于RedisStore，我们可以直接调用Verify方法
	// 这里为了保持接口一致性，仍然使用标准的Verify方法
	return c.store.Verify(id, answer, true)
}
