package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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

	if err := companyRepo.Create(ctx, company); err != nil {
		fmt.Printf("⚠️  Company might already exist: %v\n", err)
		// Try to get existing company by its ID (which we set)
		existing, _ := companyRepo.GetByID(ctx, company.ID)
		if existing != nil {
			company = existing
			fmt.Printf("📂 Using existing company: %s (%s)\n", company.Name, company.ID)
		} else {
			fmt.Println("⚠️  Could not find or create company. Continuing with generated ID...")
		}
	} else {
		fmt.Printf("✅ Created Company: %s (%s)\n", company.Name, company.ID)
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
	}

	var ownerUser *entity.User
	for _, u := range users {
		user, err := entity.NewUser(company.ID, u.Email, u.Password, u.Name, u.Role)
		if err != nil {
			log.Printf("⚠️  Failed to create user entity for %s: %v", u.Email, err)
			continue
		}

		if err := userRepo.Create(ctx, user); err != nil {
			fmt.Printf("⚠️  User %s might already exist: %v\n", u.Email, err)
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
	fmt.Printf("   • Users: 4 (owner, admin, accountant, viewer)\n")
	fmt.Printf("   • Accounts: %d\n", len(accounts))
	fmt.Printf("   • Periods: %d\n", len(periods))
	fmt.Printf("   • Journals: 5 (4 posted, 1 draft)\n")
	fmt.Println("\n📌 Login credentials:")
	fmt.Println("   • owner@test.com / Password123 (OWNER)")
	fmt.Println("   • admin@test.com / Password123 (ADMIN)")
	fmt.Println("   • accountant@test.com / Password123 (ACCOUNTANT)")
	fmt.Println("   • viewer@test.com / Password123 (VIEWER)")
}
