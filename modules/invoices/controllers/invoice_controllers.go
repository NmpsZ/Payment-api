package controllers

import (
	"net/http"
	"strconv"

	"payment-backend/modules/invoices/usecases"
	"payment-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type InvoiceController struct {
	usecase *usecases.InvoiceUsecase
}

func NewInvoiceController(u *usecases.InvoiceUsecase) *InvoiceController {
	return &InvoiceController{usecase: u}
}

type createInvoiceItemJSON struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type createInvoiceJSON struct {
	UnitNumber string                  `json:"unit_number"`
	DueDate    string                  `json:"due_date"`
	Items      []createInvoiceItemJSON `json:"items"`
}

func (ctrl *InvoiceController) Create(c *gin.Context) error {
	var body createInvoiceJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		return utils.ErrBadRequest("invalid request body")
	}

	items := make([]usecases.CreateInvoiceItem, len(body.Items))
	for i, it := range body.Items {
		items[i] = usecases.CreateInvoiceItem{
			Description: it.Description,
			Amount:      it.Amount,
		}
	}

	resp, err := ctrl.usecase.Create(c.Request.Context(), usecases.CreateInvoiceRequest{
		UnitNumber: body.UnitNumber,
		DueDate:    body.DueDate,
		Items:      items,
	})
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, resp)
	return nil
}

func (ctrl *InvoiceController) GetByID(c *gin.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return utils.ErrBadRequest("invoice id must be a positive integer")
	}

	resp, ucErr := ctrl.usecase.GetByID(c.Request.Context(), id)
	if ucErr != nil {
		return ucErr
	}
	c.JSON(http.StatusOK, resp)
	return nil
}
