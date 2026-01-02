package http

import (
	"github.com/gin-gonic/gin"
	"github.com/herman-xphp/bukuo/internal/delivery/http/handler"
	"github.com/herman-xphp/bukuo/internal/delivery/http/middleware"
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
	}

	// Protected API routes
	api := r.Group("/api")
	api.Use(authMW.Authenticate())
	{
		// Auth
		api.GET("/me", h.Auth.Me)

		// Users (Admin only)
		users := api.Group("/users")
		users.Use(authMW.RequireRole("ADMIN"))
		{
			users.POST("", h.User.Create)
			users.GET("", h.User.List)
			users.DELETE("/:id", h.User.Delete)
		}

		// Accounts
		accounts := api.Group("/accounts")
		{
			accounts.GET("", h.Account.GetAll)
			accounts.GET("/:id", h.Account.GetByID)

			// Accountants and Admins can modify
			protected := accounts.Group("")
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
			{
				protected.POST("/stock-in", h.Inventory.StockIn)
				protected.POST("/stock-out", h.Inventory.StockOut)
			}
		}

		// Periods (Admin/Accountant)
		periods := api.Group("/periods")
		{
			periods.GET("", h.Period.GetAll)
			periods.GET("/:id", h.Period.GetByID)

			protected := periods.Group("")
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
			protected.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
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
		closingRoutes.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
		{
			closingRoutes.GET("/preview/:period_id", h.Closing.PreviewClosing)
			closingRoutes.POST("/period", h.Closing.ClosePeriod)
		}

		// Opening Balance (Admin/Accountant)
		openingRoutes := api.Group("/opening-balance")
		openingRoutes.Use(authMW.RequireRole("ADMIN", "ACCOUNTANT"))
		{
			openingRoutes.GET("/template", h.Opening.Template)
			openingRoutes.POST("/import", h.Opening.Import)
		}
	}
}
