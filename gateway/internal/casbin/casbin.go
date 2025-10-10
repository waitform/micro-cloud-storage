package casbin

import (
	"log"
	"sync"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/waitform/micro-cloud-storage/global"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const configPath = "/home/haobin/桌面/test/go_test/micro-cloud-storage/gateway/config/rbac_model.conf"

var (
	enforcer *casbin.Enforcer
	once     sync.Once
)

// Init 初始化 Casbin，只调用一次
func InitCasbin() {
	g, err := global.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	dsn := g.Database.DSN
	once.Do(func() {
		m, err := model.NewModelFromFile(configPath)
		if err != nil {
			log.Fatal(err)
		}

		// 连接数据库

		dsn := dsn
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatal(err)
		}

		// 使用GORM适配器
		adapter, err := gormadapter.NewAdapterByDB(db)
		if err != nil {
			log.Fatal(err)
		}

		// 创建执行器
		enforcer, err = casbin.NewEnforcer(m, adapter)
		if err != nil {
			log.Fatal(err)
		}

		// 加载策略
		err = enforcer.LoadPolicy()
		if err != nil {
			log.Fatal(err)
		}
	})
}

// GetEnforcer 提供全局访问接口
func GetEnforcer() *casbin.Enforcer {
	if enforcer == nil {
		InitCasbin()
	}
	return enforcer
}

// AddPolicy 添加策略
func AddPolicy(userID, resource, action string) (bool, error) {
	return enforcer.AddPolicy(userID, resource, action)
}

// RemovePolicy 删除策略
func RemovePolicy(userID, resource, action string) (bool, error) {
	return enforcer.RemovePolicy(userID, resource, action)
}

// AddRoleForUser 为用户添加角色
func AddRoleForUser(user, role string) (bool, error) {
	return enforcer.AddRoleForUser(user, role)
}

// DeleteRoleForUser 删除用户角色
func DeleteRoleForUser(user, role string) (bool, error) {
	return enforcer.DeleteRoleForUser(user, role)
}

// ReloadPolicy 重新加载策略
func ReloadPolicy() error {
	return enforcer.LoadPolicy()
}
