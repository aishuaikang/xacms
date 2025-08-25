package services

import (
	"errors"
	"sort"
	"xacms/internal/server"
	"xacms/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CommonService 公共服务接口
type CommonService interface {
	GetItems(model any) error
	GetItemByID(id uuid.UUID, model any) error
	DeleteItemByID(model any, id uuid.UUID) error
	ValidateBody(c *fiber.Ctx, model any) error
	ValidateQuery(c *fiber.Ctx, model any) error
	GetAPIs() []fiber.Route
}

// commonService 公共服务实现
type commonService struct {
	db          *gorm.DB
	validator   *utils.ValidationMiddleware
	fiberServer *server.FiberServer
}

// NewCommonService 创建公共服务实例
func NewCommonService(db *gorm.DB, validator *utils.ValidationMiddleware, fiberServer *server.FiberServer) CommonService {
	return &commonService{
		db:          db,
		validator:   validator,
		fiberServer: fiberServer,
	}
}

// GetItems 获取多个数据
func (s *commonService) GetItems(model any) error {
	// 如果有 order 字段则对 order 字段进行排序，其次在进行创建时间排序
	if err := s.db.Order("`order` ASC, created_at DESC").Find(model).Error; err != nil {
		return err
	}
	return nil
}

// GetItemByID 根据ID获取单个数据
func (s *commonService) GetItemByID(id uuid.UUID, model any) error {
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

// GetAPIs 获取API列表
func (s *commonService) GetAPIs() []fiber.Route {
	routeMap := make(map[string][]fiber.Route) // 键: 路径+名称, 值: 具有相同路径+名称的路由

	allroutes := s.fiberServer.GetRoutes(true)

	// 按路径+名称分组路由
	for _, route := range allroutes {
		key := route.Path + "|" + route.Name
		routeMap[key] = append(routeMap[key], route)
	}

	var result []fiber.Route
	// 处理每个分组
	for _, routes := range routeMap {
		if len(routes) == 1 {
			// 只有一个路由，无论方法如何都保留它
			result = append(result, routes[0])

		} else {
			// 具有相同路径+名称的多个路由
			hasNonHead := false
			var headRoute *fiber.Route

			for i := range routes {
				if routes[i].Method == fiber.MethodHead {
					if headRoute == nil {
						headRoute = &routes[i]
					}
				} else {
					hasNonHead = true
					result = append(result, routes[i])
				}
			}

			// 如果没有找到非HEAD路由，保留HEAD路由
			if !hasNonHead && headRoute != nil {
				result = append(result, *headRoute)
			}
		}
	}

	// 排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}
