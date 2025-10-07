# Postman 测试脚本使用说明

本文档介绍了如何使用提供的 Postman 测试脚本来测试云存储系统的 API。

## 文件说明

1. `CloudStorage API.postman_collection.json` - Postman 测试集合文件
2. `CloudStorage API.postman_environment.json` - Postman 环境配置文件

## 导入步骤

1. 打开 Postman 应用程序
2. 点击左上角的 "Import" 按钮
3. 选择 "Upload Files" 选项
4. 选择 `CloudStorage API.postman_collection.json` 文件并导入
5. 再次点击 "Import" 按钮
6. 选择 `CloudStorage API.postman_environment.json` 文件并导入

## 环境配置

导入后，您需要配置环境变量：

1. 在 Postman 中，点击右上角的眼睛图标（Environment Quick Look）
2. 选择 "Cloud Storage API Environment"
3. 点击编辑按钮（铅笔图标）
4. 根据您的实际部署情况修改以下变量：
   - `base_url`: 网关服务的基地址（默认为 http://localhost:8080）
   - `token`: JWT 认证令牌（在登录后需要手动设置）

## API 测试流程

### 1. 用户注册和登录

1. 首先运行 "User > Register" 请求注册一个新用户
2. 然后运行 "User > Login" 请求进行登录
3. 从登录响应中复制 `token` 字段的值
4. 更新环境变量中的 `token` 值

### 2. 文件相关操作

1. 运行 "File > Init Upload" 初始化文件上传
2. 运行 "File > Upload Part" 上传文件分片
3. 运行 "File > Complete Upload" 完成文件上传
4. 运行 "File > Get File Info" 获取文件信息
5. 运行 "File > Generate Presigned URL" 生成预签名URL

### 3. 分享相关操作

1. 运行 "Share > Create Share" 创建分享
2. 运行 "Share > Get Share Info" 获取分享信息
3. 运行 "Share > Validate Access" 验证访问权限

## 注意事项

1. 需要认证的接口（标记有 Authorization 头）会自动使用环境变量中的 token
2. 在运行测试前，请确保所有相关服务都已启动
3. 某些接口需要前置条件，如上传文件前需要先注册用户并登录
4. 测试数据（如 user_id, file_id 等）可能需要根据实际情况进行调整

## API 接口列表

### 用户相关接口

- `POST /api/user/register` - 用户注册
- `POST /api/user/login` - 用户登录
- `GET /api/user/info` - 获取用户信息（需要认证）

### 文件相关接口

- `POST /api/file/upload/init` - 初始化文件上传（需要认证）
- `POST /api/file/upload/part` - 上传文件分片（需要认证）
- `POST /api/file/upload/ complete` - 完成文件上传（需要认证）
- `GET /api/file/info` - 获取文件信息（需要认证）
- `POST /api/file/presigned-url` - 生成预签名URL（需要认证）

### 分享相关接口

- `POST /api/share/create` - 创建分享（需要认证）
- `GET /api/share/info` - 获取分享信息
- `POST /api/share/validate` - 验证访问权限