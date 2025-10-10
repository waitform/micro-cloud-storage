package casbin

import (
	"testing"

	"github.com/waitform/micro-cloud-storage/internal/casbin"
)

func TestCasbinPermissions(t *testing.T) {
	// 初始化Casbin

	casbin.InitCasbin()

	// 获取执行器
	enforcer := casbin.GetEnforcer()

	// 清除可能存在的旧策略（测试前清理）
	enforcer.ClearPolicy()

	// 测试1: 添加策略并检查权限
	t.Run("AddPolicyAndCheckPermission", func(t *testing.T) {
		// 添加用户1对文件100的读权限
		added, err := casbin.AddPolicy("1", "file:100", "read")
		if err != nil {
			t.Fatalf("添加策略失败: %v", err)
		}
		if !added {
			t.Log("策略可能已存在")
		}

		// 检查用户1是否具有对文件100的读权限
		allowed, err := enforcer.Enforce("1", "file:100", "read")
		if err != nil {
			t.Fatalf("检查权限时出错: %v", err)
		}
		if !allowed {
			t.Error("期望用户1具有对文件100的读权限，但被拒绝")
		}

		// 检查用户1是否具有对文件100的写权限（应该没有）
		allowed, err = enforcer.Enforce("1", "file:100", "write")
		if err != nil {
			t.Fatalf("检查权限时出错: %v", err)
		}
		if allowed {
			t.Error("期望用户1不具有对文件100的写权限，但被允许")
		}
	})

	// 测试2: 添加更多权限
	t.Run("AddMorePolicies", func(t *testing.T) {
		// 添加用户1对文件100的写权限
		_, err := casbin.AddPolicy("1", "file:100", "write")
		if err != nil {
			t.Fatalf("添加写权限失败: %v", err)
		}

		// 添加用户2对文件100的读权限
		_, err = casbin.AddPolicy("2", "file:100", "read")
		if err != nil {
			t.Fatalf("添加用户2读权限失败: %v", err)
		}

		// 验证用户1现在有了写权限
		allowed, err := enforcer.Enforce("1", "file:100", "write")
		if err != nil {
			t.Fatalf("检查写权限时出错: %v", err)
		}
		if !allowed {
			t.Error("期望用户1具有对文件100的写权限，但被拒绝")
		}

		// 验证用户2有读权限但没有写权限
		allowed, err = enforcer.Enforce("2", "file:100", "read")
		if err != nil {
			t.Fatalf("检查读权限时出错: %v", err)
		}
		if !allowed {
			t.Error("期望用户2具有对文件100的读权限，但被拒绝")
		}

		allowed, err = enforcer.Enforce("2", "file:100", "write")
		if err != nil {
			t.Fatalf("检查写权限时出错: %v", err)
		}
		if allowed {
			t.Error("期望用户2不具有对文件100的写权限，但被允许")
		}
	})

	// 测试3: 删除策略
	t.Run("RemovePolicy", func(t *testing.T) {
		// 删除用户1的读权限
		removed, err := casbin.RemovePolicy("1", "file:100", "read")
		if err != nil {
			t.Fatalf("删除策略失败: %v", err)
		}
		if !removed {
			t.Log("策略可能不存在")
		}

		// 验证用户1不再有读权限
		allowed, err := enforcer.Enforce("1", "file:100", "read")
		if err != nil {
			t.Fatalf("检查读权限时出错: %v", err)
		}
		if allowed {
			t.Error("期望用户1不再具有对文件100的读权限，但被允许")
		}

		// 验证用户1仍有写权限
		allowed, err = enforcer.Enforce("1", "file:100", "write")
		if err != nil {
			t.Fatalf("检查写权限时出错: %v", err)
		}
		if !allowed {
			t.Error("期望用户1仍具有对文件100的写权限，但被拒绝")
		}
	})

	// 测试4: 角色相关功能
	t.Run("RoleManagement", func(t *testing.T) {
		// 添加用户到角色
		added, err := casbin.AddRoleForUser("1", "admin")
		if err != nil {
			t.Fatalf("添加角色失败: %v", err)
		}
		if !added {
			t.Log("角色关系可能已存在")
		}

		// 给角色添加权限
		_, err = casbin.AddPolicy("admin", "file:*", "read")
		if err != nil {
			t.Fatalf("给角色添加权限失败: %v", err)
		}

		// 检查用户是否继承了角色权限
		allowed, err := enforcer.Enforce("1", "file:200", "read")
		if err != nil {
			t.Fatalf("检查继承权限时出错: %v", err)
		}
		if !allowed {
			t.Error("期望用户1通过admin角色具有对任何文件的读权限，但被拒绝")
		}

		// 删除用户的角色
		deleted, err := casbin.DeleteRoleForUser("1", "admin")
		if err != nil {
			t.Fatalf("删除角色失败: %v", err)
		}
		if !deleted {
			t.Log("角色关系可能不存在")
		}

		// 验证用户不再具有继承的权限
		allowed, err = enforcer.Enforce("1", "file:200", "read")
		if err != nil {
			t.Fatalf("检查继承权限时出错: %v", err)
		}
		if allowed {
			t.Error("期望用户1不再具有通过角色继承的权限，但被允许")
		}
	})
}
