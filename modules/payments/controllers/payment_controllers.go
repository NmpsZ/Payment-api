package controllers

import (
	"net/http"

	"payment-backend/modules/payments/usecases"
	"payment-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

type PaymentController struct {
	usecase *usecases.PaymentUsecase
}

func NewPaymentController(u *usecases.PaymentUsecase) *PaymentController {
	return &PaymentController{usecase: u}
}

type createPaymentJSON struct {
	UnitNumber string  `json:"unit_number"`
	Amount     float64 `json:"amount"`
}

func (ctrl *PaymentController) Create(c *gin.Context) error {
	var body createPaymentJSON
	if err := c.ShouldBindJSON(&body); err != nil {
		return utils.ErrBadRequest("invalid request body")
	}

	resp, err := ctrl.usecase.PostPayment(c.Request.Context(), usecases.PostPaymentRequest{
		UnitNumber: body.UnitNumber,
		Amount:     body.Amount,
	})
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, resp)
	return nil
}
