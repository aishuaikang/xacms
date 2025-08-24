package services

import (
	"errors"
	"new-spbatc-drone-platform/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommonService 公共服务接口
type CommonService interface {
	GetItemByID(model any, id uuid.UUID) error
	DeleteItemByID(model any, id uuid.UUID) error
	ValidateBody(c *fiber.Ctx, model any) error
	ValidateQuery(c *fiber.Ctx, model any) error
}

// commonService 公共服务实现
type commonService struct {
	db        *gorm.DB
	validator *utils.ValidationMiddleware
}

// NewCommonService 创建公共服务实例
func NewCommonService(db *gorm.DB, validator *utils.ValidationMiddleware) CommonService {
	return &commonService{
		db:        db,
		validator: validator,
	}
}

// GetItemByID 根据ID获取单个数据
func (s *commonService) GetItemByID(model any, id uuid.UUID) error {
	if err := s.db.First(model, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// DeleteItemByID 根据ID删除单个数据
func (s *commonService) DeleteItemByID(model any, id uuid.UUID) error {
	if err := s.db.Delete(model, "id = ?", id).Error; err != nil {
		return err
	}
	return nil
}

// ValidateBody 验证请求体
func (s *commonService) ValidateBody(c *fiber.Ctx, model any) error {
	// 解析请求体
	if err := c.BodyParser(model); err != nil {
		return errors.New("请求体格式错误")
	}

	// 验证请求数据
	if errs := s.validator.ValidateStruct(model); len(errs) > 0 {
		return errors.New(errs[0])
	}
	return nil
}

// ValidateQuery 验证查询参数
func (s *commonService) ValidateQuery(c *fiber.Ctx, model any) error {
	// 解析查询参数
	if err := c.QueryParser(model); err != nil {
		return errors.New("查询参数格式错误")
	}

	// 验证查询数据
	if errs := s.validator.ValidateStruct(model); len(errs) > 0 {
		return errors.New(errs[0])
	}
	return nil
}
