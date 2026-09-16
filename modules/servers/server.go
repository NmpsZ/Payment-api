package servers

import (
	invoiceCtrl "payment-backend/modules/invoices/controllers"
	invoiceRepo "payment-backend/modules/invoices/repositories"
	invoiceUC "payment-backend/modules/invoices/usecases"
	paymentCtrl "payment-backend/modules/payments/controllers"
	paymentRepo "payment-backend/modules/payments/repositories"
	paymentUC "payment-backend/modules/payments/usecases"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

func NewRouter(db *sqlx.DB) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	invRepo := invoiceRepo.NewInvoiceRepository(db)
	invUC := invoiceUC.NewInvoiceUsecase(invRepo)
	inv := invoiceCtrl.NewInvoiceController(invUC)

	r.POST("/invoices", handle(inv.Create))
	r.GET("/invoices/:id", handle(inv.GetByID))

	payRepo := paymentRepo.NewPaymentRepository(db)
	payUC := paymentUC.NewPaymentUsecase(payRepo)
	pay := paymentCtrl.NewPaymentController(payUC)

	r.POST("/payments", handle(pay.Create))

	return r
}
