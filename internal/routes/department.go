package routes

import (
	"new-spbatc-drone-platform/internal/routes/dto"
	"new-spbatc-drone-platform/internal/services"
	"new-spbatc-drone-platform/internal/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// DepartmentHandler 部门处理器
type DepartmentHandler struct {
	Validator         *utils.ValidationMiddleware
	DepartmentService services.DepartmentService
}

// NewDepartmentHandler 创建部门处理器
func NewDepartmentHandler(validator *utils.ValidationMiddleware, departmentService services.DepartmentService) *DepartmentHandler {
	return &DepartmentHandler{
		Validator:         validator,
		DepartmentService: departmentService,
	}
}

// RegisterRoutes 注册部门相关路由
func (h *DepartmentHandler) RegisterRoutes(router fiber.Router) {
	deptGroup := router.Group("/departments")

	deptGroup.Get("/", h.GetDepartments)
	deptGroup.Post("/", h.CreateDepartment)
	deptGroup.Get("/:id", h.GetDepartment)
	deptGroup.Put("/:id", h.UpdateDepartment)
	deptGroup.Delete("/:id", h.DeleteDepartment)
	deptGroup.Get("/tree", h.GetDepartmentTree)
	deptGroup.Get("/:id/users", h.GetDepartmentUsers)
}

// GetDepartments 获取部门列表
func (h *DepartmentHandler) GetDepartments(c *fiber.Ctx) error {
	// TODO: 实现获取部门列表逻辑
	departments := []map[string]interface{}{
		{"id": 1, "name": "IT Department", "parent_id": nil},
		{"id": 2, "name": "HR Department", "parent_id": nil},
	}

	return c.JSON(dto.SuccessResponse(departments))
}

// CreateDepartment 创建部门
func (h *DepartmentHandler) CreateDepartment(c *fiber.Ctx) error {
	var deptData map[string]interface{}
	if err := c.BodyParser(&deptData); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid request body"))
	}

	// TODO: 实现创建部门逻辑
	deptData["id"] = 3

	return c.Status(201).JSON(dto.SuccessResponse(deptData))
}

// GetDepartment 获取单个部门
func (h *DepartmentHandler) GetDepartment(c *fiber.Ctx) error {
	id := c.Params("id")
	deptID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid department ID"))
	}

	// TODO: 从数据库获取部门
	department := map[string]interface{}{
		"id":        deptID,
		"name":      "Department " + id,
		"parent_id": nil,
	}

	return c.JSON(dto.SuccessResponse(department))
}

// UpdateDepartment 更新部门
func (h *DepartmentHandler) UpdateDepartment(c *fiber.Ctx) error {
	id := c.Params("id")
	deptID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid department ID"))
	}

	var deptData map[string]interface{}
	if err := c.BodyParser(&deptData); err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid request body"))
	}

	// TODO: 实现更新部门逻辑
	deptData["id"] = deptID

	return c.JSON(dto.SuccessResponse(deptData))
}

// DeleteDepartment 删除部门
func (h *DepartmentHandler) DeleteDepartment(c *fiber.Ctx) error {
	id := c.Params("id")
	deptID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid department ID"))
	}

	// TODO: 实现删除部门逻辑

	return c.JSON(dto.SuccessResponse(map[string]interface{}{
		"message": "Department deleted successfully",
		"id":      deptID,
	}))
}

// GetDepartmentTree 获取部门树结构
func (h *DepartmentHandler) GetDepartmentTree(c *fiber.Ctx) error {
	// TODO: 实现获取部门树逻辑
	deptTree := []map[string]interface{}{
		{
			"id":   1,
			"name": "IT Department",
			"children": []map[string]interface{}{
				{"id": 3, "name": "Development Team"},
				{"id": 4, "name": "QA Team"},
			},
		},
		{
			"id":       2,
			"name":     "HR Department",
			"children": []map[string]interface{}{},
		},
	}

	return c.JSON(dto.SuccessResponse(deptTree))
}

// GetDepartmentUsers 获取部门下的用户
func (h *DepartmentHandler) GetDepartmentUsers(c *fiber.Ctx) error {
	id := c.Params("id")
	deptID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(400).JSON(dto.ErrorResponse(400, "Invalid department ID"))
	}

	// TODO: 实现获取部门用户逻辑
	users := []map[string]interface{}{
		{"id": 1, "name": "John Doe", "email": "john@example.com", "department_id": deptID},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com", "department_id": deptID},
	}

	return c.JSON(dto.SuccessResponse(users))
}
