package http

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/delivery/http/middleware"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	Health       *handler.HealthHandler
	Auth         *handler.AuthHandler
	Journal      *handler.JournalHandler
	Account      *handler.AccountHandler
	Period       *handler.PeriodHandler
	Report       *handler.ReportHandler
	Closing      *handler.ClosingHandler
	Opening      *handler.OpeningHandler
	User         *handler.UserHandler
	Contact      *handler.ContactHandler
	Unit         *handler.UnitHandler
	Category     *handler.CategoryHandler
	Product      *handler.ProductHandler
	Currency     *handler.CurrencyHandler
	ExchangeRate *handler.ExchangeRateHandler
	Warehouse    *handler.WarehouseHandler
	Inventory    *handler.InventoryHandler
	Sales        *handler.SalesHandler
	Upload       *handler.UploadHandler
	Quotation    *handler.QuotationHandler
	Order        *handler.OrderHandler
	Delivery     *handler.DeliveryHandler
	Opname       *handler.OpnameHandler
	Transfer     *handler.TransferHandler
	FixedAsset   *handler.FixedAssetHandler
	Tax          *handler.TaxHandler
	Banking      *handler.BankingHandler
	Purchasing   *handler.PurchasingHandler
	ARAP         *handler.ARAPHandler
}

// SetupRouter configures all routes
func SetupRouter(r *gin.Engine, h *Handlers, authMW *middleware.AuthMiddleware) {
	// Public routes
	r.GET("/health", h.Health.Health)
	r.GET("/ping", h.Health.Ping)

	// Auth routes (public)
	auth := r.Group("/auth")
	{
		auth.POST("/register", h.Auth.Register)
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.RefreshToken)
	}

	// Static file serving for uploads (with security headers)
	r.GET("/uploads/*filepath", handler.SecureStaticHandler("./uploads"))

	// Protected API routes
	api := r.Group("/api")
	api.Use(authMW.Authenticate())
	{
		// Auth
		api.GET("/me", h.Auth.Me)
		api.PUT("/me/profile", h.Auth.UpdateProfile)
		api.PUT("/me/password", h.Auth.ChangePassword)
		api.POST("/auth/unlock", h.Auth.Unlock)

		// File Upload
		api.POST("/upload", h.Upload.Upload)

		// Users (Admin only)
		users := api.Group("/users")
		users.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner)))
		{
			users.POST("", h.User.Create)
			users.GET("", h.User.List)
			users.GET("/:id", h.User.GetByID)
			users.PUT("/:id", h.User.Update)
			users.DELETE("/:id", h.User.Delete)
			users.POST("/:id/reset-password", h.User.ResetPassword)
			users.PUT("/:id/pin", h.User.SetPin)
		}

		// Accounts
		accounts := api.Group("/accounts")
		{
			accounts.GET("", h.Account.GetAll)
			accounts.GET("/:id", h.Account.GetByID)

			// Accountants and Admins can modify
			protected := accounts.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Account.Create)
				protected.PUT("/:id", h.Account.Update)
				protected.DELETE("/:id", h.Account.Delete)
			}
		}

		// Contacts (Customers/Suppliers)
		contacts := api.Group("/contacts")
		{
			contacts.GET("", h.Contact.List)
			contacts.GET("/:id", h.Contact.GetByID)

			protected := contacts.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Contact.Create)
				protected.PUT("/:id", h.Contact.Update)
				protected.DELETE("/:id", h.Contact.Delete)
			}
		}

		// Units of Measure
		units := api.Group("/units")
		{
			units.GET("", h.Unit.List)
			units.GET("/:id", h.Unit.GetByID)

			protected := units.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Unit.Create)
				protected.DELETE("/:id", h.Unit.Delete)
			}
		}

		// Product Categories
		categories := api.Group("/product-categories")
		{
			categories.GET("", h.Category.List)
			categories.GET("/:id", h.Category.GetByID)

			protected := categories.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Category.Create)
				protected.PUT("/:id", h.Category.Update)
				protected.DELETE("/:id", h.Category.Delete)
			}
		}

		// Products
		products := api.Group("/products")
		{
			products.GET("", h.Product.List)
			products.GET("/:id", h.Product.GetByID)

			protected := products.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Product.Create)
				protected.PUT("/:id", h.Product.Update)
				protected.DELETE("/:id", h.Product.Delete)
			}
		}

		// Currencies
		currencies := api.Group("/currencies")
		{
			currencies.GET("", h.Currency.List)
			currencies.GET("/base", h.Currency.GetBase)
			currencies.GET("/:id", h.Currency.GetByID)

			protected := currencies.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Currency.Create)
				protected.POST("/preset", h.Currency.CreateFromPreset)
				protected.PUT("/:id", h.Currency.Update)
				protected.POST("/:id/set-base", h.Currency.SetBase)
				protected.DELETE("/:id", h.Currency.Delete)
			}
		}

		// Exchange Rates
		exchangeRates := api.Group("/exchange-rates")
		{
			exchangeRates.GET("", h.ExchangeRate.List)
			exchangeRates.GET("/convert", h.ExchangeRate.Convert)
			exchangeRates.GET("/:id", h.ExchangeRate.GetByID)

			protected := exchangeRates.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.ExchangeRate.Create)
				protected.PUT("/:id", h.ExchangeRate.Update)
				protected.DELETE("/:id", h.ExchangeRate.Delete)
			}
		}

		// Warehouses
		warehouses := api.Group("/warehouses")
		{
			warehouses.GET("", h.Warehouse.List)
			warehouses.GET("/:id", h.Warehouse.GetByID)

			protected := warehouses.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Warehouse.Create)
				protected.PUT("/:id", h.Warehouse.Update)
				protected.POST("/:id/set-default", h.Warehouse.SetDefault)
				protected.DELETE("/:id", h.Warehouse.Delete)
			}
		}

		// Inventory
		inventory := api.Group("/inventory")
		{
			inventory.GET("/stock", h.Inventory.GetStock)
			inventory.GET("/transactions", h.Inventory.ListTransactions)

			protected := inventory.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("/stock-in", h.Inventory.StockIn)
				protected.POST("/stock-out", h.Inventory.StockOut)

				// Stock Opname
				protected.GET("/opname", h.Opname.List)
				protected.GET("/opname/:id", h.Opname.Get)
				protected.POST("/opname", h.Opname.Create)
				protected.POST("/opname/:id/approve", h.Opname.Approve)

				// Stock Transfers
				protected.GET("/transfers", h.Transfer.List)
				protected.GET("/transfers/:id", h.Transfer.Get)
				protected.POST("/transfers", h.Transfer.Create)
			}
		}

		// Sales
		salesGroup := api.Group("/sales")
		{
			salesGroup.GET("/invoices", h.Sales.ListInvoices)
			salesGroup.GET("/invoices/:id", h.Sales.GetInvoice)

			protected := salesGroup.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("/invoices", h.Sales.CreateInvoice)
				protected.POST("/invoices/:id/void", h.Sales.VoidInvoice)
			}
		}

		// Quotations (Sales Cycle)
		quotations := api.Group("/quotations")
		{
			quotations.GET("", h.Quotation.List)
			quotations.GET("/:id", h.Quotation.Get)

			protected := quotations.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Quotation.Create)
				protected.POST("/:id/send", h.Quotation.Send)
				protected.POST("/:id/accept", h.Quotation.Accept)
				protected.DELETE("/:id", h.Quotation.Delete)
			}
		}

		// Sales Orders (Sales Cycle)
		orders := api.Group("/orders")
		{
			orders.GET("", h.Order.List)
			orders.GET("/:id", h.Order.Get)

			protected := orders.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Order.Create)
				protected.POST("/:id/confirm", h.Order.Confirm)
				protected.POST("/:id/cancel", h.Order.Cancel)
			}
		}

		// Delivery Orders (Sales Cycle)
		deliveries := api.Group("/deliveries")
		{
			deliveries.GET("", h.Delivery.List)
			deliveries.GET("/:id", h.Delivery.Get)

			protected := deliveries.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Delivery.Create)
			}
		}

		// Periods (Admin/Accountant)
		periods := api.Group("/periods")
		{
			periods.GET("", h.Period.GetAll)
			periods.GET("/:id", h.Period.GetByID)

			protected := periods.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Period.Create)
				protected.PUT("/:id", h.Period.Update)
				protected.DELETE("/:id", h.Period.Delete)
				protected.POST("/:id/close", h.Period.Close)
			}
		}

		// Journals
		journals := api.Group("/journals")
		{
			journals.GET("", h.Journal.List)                     // List all journals
			journals.GET("/pending", h.Journal.PendingApprovals) // List pending approvals
			journals.GET("/:id", h.Journal.GetByID)

			protected := journals.Group("")
			protected.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
			{
				protected.POST("", h.Journal.Create)
				protected.PUT("/:id", h.Journal.Update)
				protected.POST("/:id/post", h.Journal.Post)
				protected.POST("/:id/reverse", h.Journal.Reverse)
				protected.POST("/:id/submit-approval", h.Journal.SubmitForApproval)
				protected.POST("/:id/approve", h.Journal.Approve)
				protected.POST("/:id/reject", h.Journal.Reject)
			}
		}

		// Reports
		reports := api.Group("/reports")
		{
			reports.GET("/dashboard", h.Report.Dashboard)
			reports.GET("/trial-balance", h.Report.TrialBalance)
			reports.GET("/ledger/:account_id", h.Report.GeneralLedger)
			reports.GET("/income-statement", h.Report.IncomeStatement)
			reports.GET("/balance-sheet", h.Report.BalanceSheet)
			reports.GET("/cash-flow", h.Report.CashFlow)
		}

		// Closing (Admin/Accountant)
		closingRoutes := api.Group("/closing")
		closingRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			closingRoutes.GET("/preview/:period_id", h.Closing.PreviewClosing)
			closingRoutes.POST("/period", h.Closing.ClosePeriod)
		}

		// Fixed Assets (Admin/Accountant)
		assets := api.Group("/assets")
		assets.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			assets.GET("", h.FixedAsset.List)
			assets.GET("/:id", h.FixedAsset.Get)
			assets.POST("", h.FixedAsset.Create)
			assets.POST("/:id/dispose", h.FixedAsset.Dispose)
			assets.POST("/depreciation", h.FixedAsset.RunDepreciation)
			assets.GET("/categories", h.FixedAsset.ListCategories)
			assets.POST("/categories", h.FixedAsset.CreateCategory)
		}

		// Tax (Admin/Accountant)
		taxRoutes := api.Group("/tax")
		taxRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			// Tax Rates
			taxRoutes.GET("/rates", h.Tax.ListRates)
			taxRoutes.GET("/rates/:id", h.Tax.GetRate)
			taxRoutes.POST("/rates", h.Tax.CreateRate)
			taxRoutes.DELETE("/rates/:id", h.Tax.DeleteRate)

			// Tax Returns (SPT)
			taxRoutes.GET("/returns", h.Tax.ListReturns)
			taxRoutes.GET("/returns/:id", h.Tax.GetReturn)
			taxRoutes.POST("/returns", h.Tax.CreateReturn)
			taxRoutes.POST("/returns/:id/file", h.Tax.FileReturn)
			taxRoutes.POST("/returns/:id/pay", h.Tax.PayReturn)
		}

		// Banking (Admin/Accountant)
		bankingRoutes := api.Group("/banking")
		bankingRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			// Bank Accounts
			bankingRoutes.GET("/accounts", h.Banking.ListAccounts)
			bankingRoutes.GET("/accounts/:id", h.Banking.GetAccount)
			bankingRoutes.POST("/accounts", h.Banking.CreateAccount)
			bankingRoutes.GET("/accounts/:id/transactions", h.Banking.ListTransactions)

			// Transactions
			bankingRoutes.POST("/transactions", h.Banking.CreateTransaction)
			bankingRoutes.POST("/transactions/:id/reconcile", h.Banking.ReconcileTransaction)
		}

		// Purchasing (Admin/Accountant)
		purchasingRoutes := api.Group("/purchasing")
		purchasingRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			// Purchase Orders
			purchasingRoutes.GET("/orders", h.Purchasing.ListOrders)
			purchasingRoutes.GET("/orders/:id", h.Purchasing.GetOrder)
			purchasingRoutes.POST("/orders", h.Purchasing.CreateOrder)
			purchasingRoutes.POST("/orders/:id/approve", h.Purchasing.ApproveOrder)

			// Purchase Invoices
			purchasingRoutes.GET("/invoices", h.Purchasing.ListInvoices)
			purchasingRoutes.GET("/invoices/:id", h.Purchasing.GetInvoice)
			purchasingRoutes.POST("/invoices", h.Purchasing.CreateInvoice)
		}

		// AR/AP Reporting (Admin/Accountant)
		arapRoutes := api.Group("/arap")
		arapRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			arapRoutes.GET("/summary", h.ARAP.Summary)
			arapRoutes.GET("/receivables/aging", h.ARAP.ARAgingReport)
			arapRoutes.GET("/receivables/outstanding", h.ARAP.OutstandingReceivables)
			arapRoutes.GET("/payables/aging", h.ARAP.APAgingReport)
			arapRoutes.GET("/payables/outstanding", h.ARAP.OutstandingPayables)
		}

		// Opening Balance (Admin/Accountant)
		openingRoutes := api.Group("/opening-balance")
		openingRoutes.Use(authMW.RequireRole(string(entity.UserRoleAdmin), string(entity.UserRoleOwner), string(entity.UserRoleAccountant)))
		{
			openingRoutes.GET("/template", h.Opening.Template)
			openingRoutes.POST("/import", h.Opening.Import)
		}
	}
}
