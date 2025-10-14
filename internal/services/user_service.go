package services

import (
	"errors"
	"uav_defender/internal/dto"
	"uav_defender/internal/models"
	"uav_defender/internal/pkg/crypto"
	"uav_defender/internal/pkg/global"
	"uav_defender/internal/pkg/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// UserService 用户服务接口
type UserService interface {
	GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[dto.UserQueryResponse], error)
	CreateUser(req dto.CreateUserRequest) (*models.User, error)
	UpdateUser(userId uint64, req dto.UpdateUserRequest) (*models.User, error)
	DeleteUser(userId uint64) error
	AssignRole(userId uint64, req dto.AssignRoleRequest) (*models.User, error)
	ChangePassword(userId uint64, req dto.ChangePasswordRequest) error
	Login(req dto.LoginRequest) (*dto.LoginResponse, error)
}

// userService 用户服务实现
type userService struct {
	db             *gorm.DB
	commonService  CommonService
	roleService    RoleService
	passwordCrypto *crypto.SimplePasswordCrypto
}

// NewUserService 创建用户服务实例
func NewUserService(db *gorm.DB, commonService CommonService, roleService RoleService) UserService {
	// 从配置文件读取盐值，或使用默认值
	passwordCrypto := crypto.NewSimplePasswordCrypto()
	return &userService{
		db:             db,
		commonService:  commonService,
		roleService:    roleService,
		passwordCrypto: passwordCrypto,
	}
}

// GetUsers 获取用户列表
func (s *userService) GetUsers(req dto.UserQueryRequest) (*dto.PaginatedResponse[dto.UserQueryResponse], error) {
	// 使用子查询统计关联数据数量，避免 N+1 查询问题
	// 同时预加载 Role 和 Tenant 信息（如果需要显示）
	query := s.db.Model(&models.User{}).
		Select(`users.*,
			(SELECT COUNT(*) FROM devices WHERE devices.user_id = users.id) as device_count`).
		Preload("Role").
		Preload("Tenant")

	var total int64
	countQuery := s.db.Model(&models.User{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	var users []dto.UserQueryResponse
	if err := paginate(query, req.Page, req.PageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, err
	}

	return &dto.PaginatedResponse[dto.UserQueryResponse]{
		Total: total,
		Items: users,
	}, nil
}

// CreateUser 创建用户
func (s *userService) CreateUser(req dto.CreateUserRequest) (*models.User, error) {
	// 处理前端哈希密码
	hashedPassword, err := s.passwordCrypto.ProcessClientPassword(req.Password)
	if err != nil {
		global.Logger.Error("处理密码失败", zap.Error(err))
		return nil, errors.New("密码格式错误")
	}

	userData := &models.User{
		Nickname: req.Nickname,
		Username: req.Username,
		Password: hashedPassword,
		RoleID:   &req.RoleID,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
	}

	if err := s.db.Create(userData).Error; err != nil {
		return nil, err
	}
	return userData, nil
}

// UpdateUser 修改用户
func (s *userService) UpdateUser(userId uint64, req dto.UpdateUserRequest) (*models.User, error) {
	var user models.User
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}

	if req.Username != nil {
		user.Username = *req.Username
	}

	if req.RoleID != nil {
		// 检查角色是否存在
		exists, err := s.roleService.IsRoleExist(*req.RoleID)
		if err != nil {
			return nil, err
		}

		if !exists {
			return nil, errors.New("角色不存在")
		}
		user.RoleID = req.RoleID
	}

	if req.Email != nil {
		user.Email = req.Email
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AssignRole 分配角色
func (s *userService) AssignRole(userId uint64, req dto.AssignRoleRequest) (*models.User, error) {
	var user models.User
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户不存在")
		}
		return nil, err
	}

	// 检查角色是否存在
	exists, err := s.roleService.IsRoleExist(req.RoleID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New("角色不存在")
	}

	user.RoleID = &req.RoleID

	if err := s.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *userService) Login(req dto.LoginRequest) (*dto.LoginResponse, error) {
	var user models.User
	// 修复:先根据用户名查找用户,并关联查询租户信息
	if err := s.db.Where("username = ?", req.Username).Preload("Tenant").First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	// 使用密码加密工具验证密码
	if !s.passwordCrypto.VerifyPassword(req.Password, user.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 判断用户是否有分配角色
	if user.RoleID == nil {
		return nil, errors.New("用户未分配角色")
	}

	// 生成JWT令牌
	token, err := utils.GenerateToken(&user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

// ChangePassword 修改用户密码
func (s *userService) ChangePassword(userId uint64, req dto.ChangePasswordRequest) error {
	var user models.User
	if err := s.commonService.GetItemByID(userId, &user); err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用户不存在")
		}
		return err
	}

	// 验证旧密码是否正确
	if !s.passwordCrypto.VerifyPassword(req.OldPassword, user.Password) {
		return errors.New("旧密码不正确")
	}

	// 处理前端哈希的新密码
	hashedNewPassword, err := s.passwordCrypto.ProcessClientPassword(req.NewPassword)
	if err != nil {
		global.Logger.Error("处理新密码失败", zap.Error(err))
		return errors.New("新密码格式错误")
	}

	// 更新密码
	user.Password = hashedNewPassword
	if err := s.db.Save(&user).Error; err != nil {
		return err
	}

	return nil
}

// DeleteUser 删除用户
func (s *userService) DeleteUser(userId uint64) error {
	// 当存在关联的设备时，禁止删除用户
	var user models.User
	if err := s.db.Preload("Devices").First(&user, userId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errors.New("用户不存在")
		}
		return err
	}

	if err := s.db.Delete(&user).Error; err != nil {
		return err
	}

	return nil
}
