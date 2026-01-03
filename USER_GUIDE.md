# Panduan Penggunaan Sistem BUKUO

Selamat datang di Panduan Penggunaan BUKUO. Dokumen ini menjelaskan langkah-langkah penggunaan sistem mulai dari persiapan data hingga proses transaksi penjualan dan stok.

## 1. Persiapan Awal (Master Data)

Sebelum memulai transaksi, pastikan data master sudah tersedia.

### A. Membuat Data Produk

1. Buka menu **Products** (`/products`).
2. Klik tombol **+ Add Product**.
3. Isi data wajib:
   - **Name**: Nama produk (contoh: "Kopi Susu Gula Aren").
   - **Code/SKU**: Kode unik produk (contoh: "K001").
   - **Unit**: Satuan (Pcs, Kg, dll).
   - **Sales Price**: Harga Jual.
   - **Purchase Price**: Harga Beli (untuk hitung profit).
4. Klik **Save**. Pastikan produk muncul di daftar.

### B. Mengatur Gudang (Optional)

1. Buka menu **Settings** > **Warehouses** (`/settings/warehouses`).
2. Secara default sudah ada "Main Warehouse".
3. Jika punya cabang/gudang lain, klik **+ Add Warehouse** untuk menambah.

---

## 2. Manajemen Stok (Inventory)

Setelah produk dibuat, stok awalnya mungkin masih 0. Gunakan fitur ini untuk mengisi stok.

### A. Menambah Stok (Stock In)

Gunakan fitur ini saat: Barang baru datang dari supplier, produksi selesai, atau stock opname awal.

1. Buka menu **Inventory** (`/inventory`).
2. Klik tombol **Stock In** (ikon panah hijau ke bawah).
3. Di jendela yang muncul:
   - **Product**: Pilih produk yang akan ditambah.
   - **Warehouse**: Pilih gudang penyimpanan (default: Main Warehouse).
   - **Quantity**: Masukkan jumlah barang.
   - **Unit Cost**: Masukkan harga beli per unit (PENTING untuk menghitung rata-rata HPP).
   - **Notes**: Catatan (misal: "Stok awal" atau "Beli dari Supplier A").
4. Klik **Confirm**.
5. Stok di tabel "Inventory" akan bertambah.

### B. Mengurangi Stok (Stock Out)

Gunakan fitur ini saat: Barang rusak, hilang, kadaluarsa, atau pemakaian sendiri (bukan penjualan).

> [!NOTE]
> Stok akan **otomatis berkurang** saat Anda melakukan transaksi penjualan di POS/Sales. Jadi, Anda **TIDAK PERLU** melakukan Stock Out manual untuk setiap penjualan.

1. Buka menu **Inventory** (`/inventory`).
2. Klik tombol **Stock Out** (ikon panah merah ke atas).
3. Isi form:
   - Pilih **Product** dan **Warehouse**.
   - Masukkan **Quantity** yang keluar.
   - Beri **Notes** (contoh: "Barang Rusak" atau "Sampel Marketing").
4. Klik **Confirm**. Stok akan berkurang.

---

## 3. Transaksi Penjualan (Point of Sales)

Ini adalah fitur utama kasir untuk melayani pembeli.

### Langkah Transaksi:

1. Buka menu **POS** (`/pos`).
2. **Pilih Produk**: Klik produk di daftar (kiri) untuk memasukkan ke keranjang (kanan).
   - Gunakan _Search Bar_ jika produk banyak.
3. **Atur Keranjang**:
   - Ubah jumlah (Qty) dengan tombol + atau -.
   - Klik ikon sampah untuk hapus item.
4. **Diskon & Pajak (Opsional)**:
   - Di bagian bawah keranjang, isi **Discount (%)** jika ada promo.
   - Isi **Tax (%)** jika ada PPN (misal 11%).
5. **Checkout**:
   - Klik tombol **Charge / Bayar**.
   - Masukkan jumlah uang yang diterima pelanggan.
   - Pilih metode bayar (Cash/Transfer - _saat ini default Cash_).
6. **Selesai**:
   - Akan muncul struk pembayaran.
   - **Print**: Klik tombol Print untuk mencetak struk thermal.
   - **Share**: Klik tombol WhatsApp untuk kirim struk digital ke pelanggan.
   - Klik "New Sale" untuk transaksi baru.

---

## 4. Riwayat & Pembatalan Transaksi

### Melihat Riwayat

1. Buka menu **Sales** (`/sales`).
2. Anda akan melihat daftar semua faktur (Invoice) yang telah dibuat.
3. Gunakan filter tanggal atau status untuk mencari transaksi.

### Membatalkan Transaksi (Void)

Jika kasir salah input:

1. Di menu **Sales**, cari transaksi yang salah.
2. Klik tombol **Tong Sampah** (Void) di kolom _Actions_ sebelah kanan.
3. Konfirmasi pembatalan.
4. Status invoice berubah jadi **CANCELLED** dan stok barang akan otomatis **dikembalikan** ke gudang.

---

## 5. Ringkasan (Dashboard)

1. Buka menu **Dashboard**.
2. Anda bisa melihat ringkasan:
   - Total Penjualan hari ini.
   - Total Profit.
   - Grafik tren penjualan.

---

**Tips:**

- Pastikan selalu input **Stock In** dengan **Unit Cost** yang benar agar laporan laba rugi akurat.
- Jika produk tidak muncul di POS atau Inventory, pastikan produk tersebut statusnya **Active** di menu Products.
