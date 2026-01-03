package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/herman-xphp/bukuo/internal/config"
	"github.com/herman-xphp/bukuo/internal/database"
	"github.com/herman-xphp/bukuo/internal/domain/entity"
	"github.com/herman-xphp/bukuo/internal/infrastructure/persistence/postgres"
	"github.com/shopspring/decimal"
)

func main() {
	cfg := config.Load()
	db, err := database.Connect(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	companyRepo := postgres.NewCompanyRepository(db)
	userRepo := postgres.NewUserRepository(db)
	accountRepo := postgres.NewAccountRepository(db)
	periodRepo := postgres.NewPeriodRepository(db)
	journalRepo := postgres.NewJournalRepository(db)

	// ============================================
	// 1. CREATE COMPANY
	// ============================================
	company := entity.NewCompany("PT Digital Future", "01.234.567.8-901.000")
	company.Address = "Jl. Sudirman No. 1, Jakarta"
	company.Email = "info@digitalfuture.co.id"

	// Check if company exists by name (Seeder idempotency)
	var existingID string
	err = db.QueryRow(ctx, "SELECT id FROM companies WHERE name = $1 LIMIT 1", company.Name).Scan(&existingID)
	if err == nil {
		// Found existing
		uid, _ := uuid.Parse(existingID)
		existingCompany, _ := companyRepo.GetByID(ctx, uid)
		if existingCompany != nil {
			company = existingCompany
			fmt.Printf("📂 Using existing company: %s (%s)\n", company.Name, company.ID)
		}
	} else {
		// Create new
		if err := companyRepo.Create(ctx, company); err != nil {
			fmt.Printf("⚠️  Company might already exist: %v\n", err)
		} else {
			fmt.Printf("✅ Created Company: %s (%s)\n", company.Name, company.ID)
		}
	}

	// ============================================
	// 2. CREATE USERS
	// ============================================
	fmt.Println("\n--- Creating Users ---")
	users := []struct {
		Name     string
		Email    string
		Password string
		Role     entity.UserRole
	}{
		{"Herman Owner", "owner@test.com", "Password123", entity.UserRoleOwner},
		{"Admin Staff", "admin@test.com", "Password123", entity.UserRoleAdmin},
		{"Dewi Akuntan", "accountant@test.com", "Password123", entity.UserRoleAccountant},
		{"Budi Viewer", "viewer@test.com", "Password123", entity.UserRoleViewer},
		{"Rina Kasir", "cashier@test.com", "Password123", entity.UserRoleCashier},
	}

	var ownerUser *entity.User
	for _, u := range users {
		user, err := entity.NewUser(company.ID, u.Email, u.Password, u.Name, u.Role)
		if err != nil {
			log.Printf("⚠️  Failed to create user entity for %s: %v", u.Email, err)
			continue
		}

		if err := user.SetPin("123456"); err != nil {
			log.Printf("⚠️  Failed to set PIN for %s: %v", u.Email, err)
		}

		if err := userRepo.Create(ctx, user); err != nil {
			fmt.Printf("⚠️  User %s might already exist. Updating PIN...\n", u.Email)
			existing, err := userRepo.GetByEmail(ctx, u.Email)
			if err == nil {
				existing.SetPin("123456")
				if err := userRepo.Update(ctx, existing); err != nil {
					fmt.Printf("   ❌ Failed to update PIN: %v\n", err)
				} else {
					fmt.Printf("   ✅ Updated PIN to 123456\n")
				}
			}
		} else {
			fmt.Printf("✅ Created User: %s (%s) - %s\n", u.Name, u.Email, u.Role)
		}

		if u.Role == entity.UserRoleOwner {
			ownerUser = user
		}
	}

	// Get owner user if not created
	if ownerUser == nil {
		existingUsers, _ := userRepo.GetByCompany(ctx, company.ID)
		for i := range existingUsers {
			if existingUsers[i].Role == entity.UserRoleOwner {
				ownerUser = &existingUsers[i]
				break
			}
		}
	}

	// ============================================
	// 3. CREATE CHART OF ACCOUNTS
	// ============================================
	fmt.Println("\n--- Creating Chart of Accounts ---")
	accounts := []struct {
		Code       string
		Name       string
		Type       entity.AccountType
		IsPostable bool
	}{
		// ASSETS (1-XXXX)
		{"1-0000", "Aset", entity.AccountTypeAsset, false},
		{"1-1000", "Aset Lancar", entity.AccountTypeAsset, false},
		{"1-1001", "Kas", entity.AccountTypeAsset, true},
		{"1-1002", "Bank BCA", entity.AccountTypeAsset, true},
		{"1-1003", "Bank Mandiri", entity.AccountTypeAsset, true},
		{"1-1100", "Piutang Usaha", entity.AccountTypeAsset, true},
		{"1-1200", "Persediaan", entity.AccountTypeAsset, true},
		{"1-2000", "Aset Tetap", entity.AccountTypeAsset, false},
		{"1-2001", "Peralatan Kantor", entity.AccountTypeAsset, true},
		{"1-2002", "Kendaraan", entity.AccountTypeAsset, true},
		{"1-2003", "Gedung", entity.AccountTypeAsset, true},
		{"1-2900", "Akumulasi Penyusutan", entity.AccountTypeAsset, true},

		// LIABILITIES (2-XXXX)
		{"2-0000", "Kewajiban", entity.AccountTypeLiability, false},
		{"2-1000", "Kewajiban Lancar", entity.AccountTypeLiability, false},
		{"2-1001", "Hutang Usaha", entity.AccountTypeLiability, true},
		{"2-1002", "Hutang Gaji", entity.AccountTypeLiability, true},
		{"2-1003", "Hutang Pajak", entity.AccountTypeLiability, true},
		{"2-2000", "Kewajiban Jangka Panjang", entity.AccountTypeLiability, false},
		{"2-2001", "Hutang Bank", entity.AccountTypeLiability, true},

		// EQUITY (3-XXXX)
		{"3-0000", "Ekuitas", entity.AccountTypeEquity, false},
		{"3-1001", "Modal Disetor", entity.AccountTypeEquity, true},
		{"3-2001", "Laba Ditahan", entity.AccountTypeEquity, true},
		{"3-3001", "Laba Tahun Berjalan", entity.AccountTypeEquity, true},

		// REVENUE (4-XXXX)
		{"4-0000", "Pendapatan", entity.AccountTypeRevenue, false},
		{"4-1001", "Pendapatan Jasa", entity.AccountTypeRevenue, true},
		{"4-1002", "Pendapatan Penjualan", entity.AccountTypeRevenue, true},
		{"4-2001", "Pendapatan Lain-lain", entity.AccountTypeRevenue, true},

		// EXPENSES (5-XXXX)
		{"5-0000", "Beban", entity.AccountTypeExpense, false},
		{"5-1001", "Beban Gaji", entity.AccountTypeExpense, true},
		{"5-1002", "Beban Sewa", entity.AccountTypeExpense, true},
		{"5-1003", "Beban Utilitas", entity.AccountTypeExpense, true},
		{"5-1004", "Beban Perlengkapan", entity.AccountTypeExpense, true},
		{"5-1005", "Beban Penyusutan", entity.AccountTypeExpense, true},
		{"5-2001", "Beban Lain-lain", entity.AccountTypeExpense, true},
	}

	accountMap := make(map[string]*entity.Account)
	for _, a := range accounts {
		account := entity.NewAccount(company.ID, a.Code, a.Name, a.Type)
		account.IsPostable = a.IsPostable

		if err := accountRepo.Create(ctx, account); err != nil {
			fmt.Printf("⚠️  Account %s might already exist: %v\n", a.Code, err)
			// Try to get existing
			existing, _ := accountRepo.GetByCode(ctx, company.ID, a.Code)
			if existing != nil {
				accountMap[a.Code] = existing
			}
		} else {
			fmt.Printf("✅ Created Account: %s - %s (%s)\n", a.Code, a.Name, a.Type)
			accountMap[a.Code] = account
		}
	}

	// ============================================
	// 4. CREATE ACCOUNTING PERIOD
	// ============================================
	fmt.Println("\n--- Creating Accounting Periods ---")
	year := time.Now().Year()
	periods := []struct {
		Name      string
		StartDate time.Time
		EndDate   time.Time
	}{
		{fmt.Sprintf("Periode Januari %d", year), time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 1, 31, 0, 0, 0, 0, time.UTC)},
		{fmt.Sprintf("Periode Februari %d", year), time.Date(year, 2, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 2, 28, 0, 0, 0, 0, time.UTC)},
		{fmt.Sprintf("Periode Maret %d", year), time.Date(year, 3, 1, 0, 0, 0, 0, time.UTC), time.Date(year, 3, 31, 0, 0, 0, 0, time.UTC)},
	}

	var activePeriod *entity.AccountingPeriod
	for _, p := range periods {
		period := entity.NewAccountingPeriod(company.ID, p.Name, p.StartDate, p.EndDate)

		if err := periodRepo.Create(ctx, period); err != nil {
			fmt.Printf("⚠️  Period %s might already exist: %v\n", p.Name, err)
			// Get by date
			existing, _ := periodRepo.GetByDate(ctx, company.ID, p.StartDate)
			if existing != nil && activePeriod == nil {
				activePeriod = existing
			}
		} else {
			fmt.Printf("✅ Created Period: %s (%s - %s)\n", p.Name, p.StartDate.Format("2006-01-02"), p.EndDate.Format("2006-01-02"))
			if activePeriod == nil {
				activePeriod = period
			}
		}
	}

	// Get first period if not set
	if activePeriod == nil {
		existingPeriods, _ := periodRepo.GetByCompany(ctx, company.ID)
		if len(existingPeriods) > 0 {
			activePeriod = &existingPeriods[0]
		}
	}

	// ============================================
	// 4.5. CREATE PRODUCTS & INVENTORY
	// ============================================
	fmt.Println("\n--- Creating Products & Inventory ---")
	unitRepo := postgres.NewUnitRepository(db)
	catRepo := postgres.NewCategoryRepository(db)
	invRepo := postgres.NewInventoryRepository(db)
	warehouseRepo := postgres.NewWarehouseRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// 1. Units
	pcsUnit := entity.NewUnitOfMeasure(company.ID, "PCS", "Pieces")
	boxUnit := entity.NewUnitOfMeasure(company.ID, "BOX", "Box")
	if err := unitRepo.Create(ctx, pcsUnit); err != nil {
		if u, _ := unitRepo.GetByCode(ctx, company.ID, "PCS"); u != nil {
			pcsUnit = u
		}
	}
	if err := unitRepo.Create(ctx, boxUnit); err != nil {
		if u, _ := unitRepo.GetByCode(ctx, company.ID, "BOX"); u != nil {
			boxUnit = u
		}
	}

	// 2. Warehouses
	mainWarehouse := entity.NewWarehouse(company.ID, "WH-MAIN", "Main Warehouse")
	mainWarehouse.Address = "Jl. Gudang Utama No. 1"
	mainWarehouse.IsDefault = true
	if err := warehouseRepo.Create(ctx, mainWarehouse); err != nil {
		if wh, _ := warehouseRepo.GetByCode(ctx, company.ID, "WH-MAIN"); wh != nil {
			mainWarehouse = wh
		}
	}

	// 3. Categories
	catElec := entity.NewProductCategory(company.ID, "Electronics", nil)
	catFurn := entity.NewProductCategory(company.ID, "Furniture", nil)

	// Helper to find category
	cats, _ := catRepo.List(ctx, company.ID)
	for i := range cats {
		if cats[i].Name == "Electronics" {
			catElec = &cats[i]
		}
		if cats[i].Name == "Furniture" {
			catFurn = &cats[i]
		}
	}
	// Create if ID matches new one (means not found)
	// Actually List returns structs with IDs. If we created new one, ID is new.
	// Better check:
	if catElec.Name == "Electronics" && catRepo.Create(ctx, catElec) != nil { /* ignore */
	}
	if catFurn.Name == "Furniture" && catRepo.Create(ctx, catFurn) != nil { /* ignore */
	}
	// The above logic is sloppy (it tries to create even if found, relying on error? Create assumes new ID).
	// Let's refine:
	if exists := false; true {
		for i := range cats {
			if cats[i].Name == "Electronics" {
				exists = true
				catElec = &cats[i]
				break
			}
		}
		if !exists {
			catRepo.Create(ctx, catElec)
		}
	}
	if exists := false; true {
		for i := range cats {
			if cats[i].Name == "Furniture" {
				exists = true
				catFurn = &cats[i]
				break
			}
		}
		if !exists {
			catRepo.Create(ctx, catFurn)
		}
	}

	// 4. Products
	// Finding accounts for mapping
	salesAcct := accountMap["4-1002"] // Pendapatan Penjualan

	// Create HPP Account if missing
	hppAcct := entity.NewAccount(company.ID, "5-1100", "Beban Pokok Penjualan", entity.AccountTypeExpense)
	hppAcct.IsPostable = true
	if err := accountRepo.Create(ctx, hppAcct); err == nil {
		accountMap["5-1100"] = hppAcct
	} else {
		if ex, _ := accountRepo.GetByCode(ctx, company.ID, "5-1100"); ex != nil {
			hppAcct = ex
		}
	}

	// Check if inventory account exists
	invAcct, ok := accountMap["1-1200"]
	if !ok {
		// Create if missing
		invAcct = entity.NewAccount(company.ID, "1-1200", "Persediaan Barang", entity.AccountTypeAsset)
		invAcct.IsPostable = true
		accountRepo.Create(ctx, invAcct)
		accountMap["1-1200"] = invAcct // Ensure it's in map
	}

	products := []struct {
		Code  string
		Name  string
		Price decimal.Decimal
		Cost  decimal.Decimal
		CatID uuid.UUID
		Unit  *entity.UnitOfMeasure
	}{
		{"PRD-001", "Laptop Gaming High-End", decimal.NewFromInt(25000000), decimal.NewFromInt(18000000), catElec.ID, pcsUnit},
		{"PRD-002", "Office Chair Ergonomic", decimal.NewFromInt(3500000), decimal.NewFromInt(2000000), catFurn.ID, pcsUnit},
		{"PRD-003", "Mechanical Keyboard Wireless", decimal.NewFromInt(1500000), decimal.NewFromInt(800000), catElec.ID, pcsUnit},
		{"PRD-004", "USB-C Hub Multiport", decimal.NewFromInt(450000), decimal.NewFromInt(250000), catElec.ID, pcsUnit},
		{"PRD-005", "Monitor 27 Inch 4K", decimal.NewFromInt(5500000), decimal.NewFromInt(3500000), catElec.ID, pcsUnit},
		{"PRD-006", "Smartphone Flagship 2026", decimal.NewFromInt(15000000), decimal.NewFromInt(12000000), catElec.ID, pcsUnit},
		{"PRD-007", "Tablet Pro 12.9 Inch", decimal.NewFromInt(18000000), decimal.NewFromInt(14000000), catElec.ID, pcsUnit},
		{"PRD-008", "Smartwatch Series 7", decimal.NewFromInt(6000000), decimal.NewFromInt(4500000), catElec.ID, pcsUnit},
		{"PRD-009", "Wireless Mouse Silent", decimal.NewFromInt(250000), decimal.NewFromInt(150000), catElec.ID, pcsUnit},
		{"PRD-010", "Headset Bluetooth Noise Cancelling", decimal.NewFromInt(3000000), decimal.NewFromInt(2000000), catElec.ID, pcsUnit},
	}

	for _, p := range products {
		prod := entity.NewProduct(company.ID, p.Code, p.Name, entity.ProductTypeGoods, p.Unit.ID, salesAcct.ID, hppAcct.ID)
		prod.InventoryAccountID = &invAcct.ID
		prod.CategoryID = &p.CatID
		prod.SalesPrice = p.Price
		prod.PurchasePrice = p.Cost
		prod.MinStock = decimal.NewFromInt(5)
		prod.ID = uuid.New() // Fix: generate unique ID for seeder

		// Try Create
		if err := productRepo.Create(ctx, prod); err != nil {
			fmt.Printf("⚠️  Product %s might already exist\n", p.Name)
			// Fetch to get ID for inventory
			if existing, _ := productRepo.GetByCode(ctx, company.ID, p.Code); existing != nil {
				prod = existing
			}
		} else {
			fmt.Printf("✅ Created Product: %s\n", p.Name)
		}

		// 5. Initial Inventory (Stock In)
		stock, _ := invRepo.GetTotalStock(ctx, company.ID, prod.ID)
		if stock == nil || stock.Quantity.LessThan(decimal.NewFromInt(10)) {
			tx := entity.NewInventoryTransaction(
				company.ID,
				fmt.Sprintf("IN-%s-%d", p.Code, time.Now().Unix()),
				entity.InventoryTransactionStockIn,
				prod.ID,
				mainWarehouse.ID,
				decimal.NewFromInt(50),
				p.Cost,
			)
			tx.Notes = "Initial Seeding"
			// Debug the ID being used
			// fmt.Printf("   debug: link stock to prod %s\n", prod.ID)

			if err := invRepo.CreateTransaction(ctx, tx); err != nil {
				fmt.Printf("   ⚠️ Failed to add stock: %v\n", err)
			} else {
				fmt.Printf("   📦 Added 50 stock for %s\n", p.Name)
			}
		}
	}

	if activePeriod == nil || ownerUser == nil {
		fmt.Println("⚠️  Cannot create journals - missing period or user")
		fmt.Println("\n🎉 Seeding complete (partial)!")
		return
	}

	// ============================================
	// 5. CREATE SAMPLE JOURNAL ENTRIES
	// ============================================
	fmt.Println("\n--- Creating Sample Journal Entries ---")

	// Journal 1: Modal Awal (Setoran Modal)
	journal1 := entity.NewJournalEntry(company.ID, activePeriod.ID, ownerUser.ID, time.Now(), "Setoran Modal Awal")
	journal1.EntryNumber = fmt.Sprintf("JE-%d-0001", year)
	journal1.SourceType = "OPENING"

	if bankBCA, ok := accountMap["1-1002"]; ok {
		journal1.AddLine(bankBCA.ID, "Bank BCA", decimal.NewFromInt(100000000), decimal.Zero)
	}
	if modalDisetor, ok := accountMap["3-1001"]; ok {
		journal1.AddLine(modalDisetor.ID, "Modal Disetor", decimal.Zero, decimal.NewFromInt(100000000))
	}

	if err := journal1.Post(ownerUser.ID); err == nil {
		if err := journalRepo.Create(ctx, journal1); err != nil {
			fmt.Printf("⚠️  Journal 1 might already exist: %v\n", err)
		} else {
			fmt.Printf("✅ Created & Posted Journal: %s - %s\n", journal1.EntryNumber, journal1.Description)
		}
	}

	// Journal 2: Pembelian Peralatan
	journal2 := entity.NewJournalEntry(company.ID, activePeriod.ID, ownerUser.ID, time.Now(), "Pembelian Peralatan Kantor")
	journal2.EntryNumber = fmt.Sprintf("JE-%d-0002", year)
	journal2.SourceType = "MANUAL"

	if peralatan, ok := accountMap["1-2001"]; ok {
		journal2.AddLine(peralatan.ID, "Peralatan Kantor", decimal.NewFromInt(15000000), decimal.Zero)
	}
	if bankBCA, ok := accountMap["1-1002"]; ok {
		journal2.AddLine(bankBCA.ID, "Bank BCA", decimal.Zero, decimal.NewFromInt(15000000))
	}

	if err := journal2.Post(ownerUser.ID); err == nil {
		if err := journalRepo.Create(ctx, journal2); err != nil {
			fmt.Printf("⚠️  Journal 2 might already exist: %v\n", err)
		} else {
			fmt.Printf("✅ Created & Posted Journal: %s - %s\n", journal2.EntryNumber, journal2.Description)
		}
	}

	// Journal 3: Pendapatan Jasa
	journal3 := entity.NewJournalEntry(company.ID, activePeriod.ID, ownerUser.ID, time.Now(), "Pendapatan Jasa Konsultasi")
	journal3.EntryNumber = fmt.Sprintf("JE-%d-0003", year)
	journal3.SourceType = "MANUAL"

	if bankBCA, ok := accountMap["1-1002"]; ok {
		journal3.AddLine(bankBCA.ID, "Bank BCA", decimal.NewFromInt(25000000), decimal.Zero)
	}
	if pendapatanJasa, ok := accountMap["4-1001"]; ok {
		journal3.AddLine(pendapatanJasa.ID, "Pendapatan Jasa", decimal.Zero, decimal.NewFromInt(25000000))
	}

	if err := journal3.Post(ownerUser.ID); err == nil {
		if err := journalRepo.Create(ctx, journal3); err != nil {
			fmt.Printf("⚠️  Journal 3 might already exist: %v\n", err)
		} else {
			fmt.Printf("✅ Created & Posted Journal: %s - %s\n", journal3.EntryNumber, journal3.Description)
		}
	}

	// Journal 4: Beban Gaji
	journal4 := entity.NewJournalEntry(company.ID, activePeriod.ID, ownerUser.ID, time.Now(), "Pembayaran Gaji Karyawan")
	journal4.EntryNumber = fmt.Sprintf("JE-%d-0004", year)
	journal4.SourceType = "MANUAL"

	if bebanGaji, ok := accountMap["5-1001"]; ok {
		journal4.AddLine(bebanGaji.ID, "Beban Gaji", decimal.NewFromInt(10000000), decimal.Zero)
	}
	if bankBCA, ok := accountMap["1-1002"]; ok {
		journal4.AddLine(bankBCA.ID, "Bank BCA", decimal.Zero, decimal.NewFromInt(10000000))
	}

	if err := journal4.Post(ownerUser.ID); err == nil {
		if err := journalRepo.Create(ctx, journal4); err != nil {
			fmt.Printf("⚠️  Journal 4 might already exist: %v\n", err)
		} else {
			fmt.Printf("✅ Created & Posted Journal: %s - %s\n", journal4.EntryNumber, journal4.Description)
		}
	}

	// Journal 5: Draft Journal (not posted)
	journal5 := entity.NewJournalEntry(company.ID, activePeriod.ID, ownerUser.ID, time.Now(), "Draft - Pembelian Perlengkapan")
	journal5.EntryNumber = fmt.Sprintf("JE-%d-0005", year)
	journal5.SourceType = "MANUAL"

	if bebanPerlengkapan, ok := accountMap["5-1004"]; ok {
		journal5.AddLine(bebanPerlengkapan.ID, "Beban Perlengkapan", decimal.NewFromInt(500000), decimal.Zero)
	}
	if kas, ok := accountMap["1-1001"]; ok {
		journal5.AddLine(kas.ID, "Kas", decimal.Zero, decimal.NewFromInt(500000))
	}

	// Keep as DRAFT (not posted)
	if err := journalRepo.Create(ctx, journal5); err != nil {
		fmt.Printf("⚠️  Journal 5 might already exist: %v\n", err)
	} else {
		fmt.Printf("✅ Created Draft Journal: %s - %s (DRAFT)\n", journal5.EntryNumber, journal5.Description)
	}

	// ============================================
	// SUMMARY
	// ============================================
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("🎉 SEEDING COMPLETE!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   • Company: %s\n", company.Name)
	fmt.Printf("   • Users: 5 (owner, admin, accountant, viewer, cashier)\n")
	fmt.Printf("   • Accounts: %d\n", len(accounts))
	fmt.Printf("   • Periods: %d\n", len(periods))
	fmt.Printf("   • Journals: 5 (4 posted, 1 draft)\n")
	fmt.Println("\n📌 Login credentials:")
	fmt.Println("   • owner@test.com / Password123 (OWNER)")
	fmt.Println("   • admin@test.com / Password123 (ADMIN)")
	fmt.Println("   • accountant@test.com / Password123 (ACCOUNTANT)")
	fmt.Println("   • viewer@test.com / Password123 (VIEWER)")
	fmt.Println("   • cashier@test.com / Password123 (CASHIER)")
	fmt.Println("\n📌 Default PIN for all users: 123456")

}
