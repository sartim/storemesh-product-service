package repository

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	productv1 "storemesh-product-service/gen/storemesh/product/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductStore struct{ db *sql.DB }

func OpenProductStore(ctx context.Context, url string) (*ProductStore, error) {
	db, err := sql.Open("pgx", url)
	if err != nil { return nil, err }
	if err := db.PingContext(ctx); err != nil { _ = db.Close(); return nil, err }
	return &ProductStore{db: db}, nil
}

func (s *ProductStore) Close() error { return s.db.Close() }

func (s *ProductStore) Insert(ctx context.Context, p *productv1.Product) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO products (id, sku, name, description, price_minor, currency, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`, p.Id, p.Sku, p.Name, p.Description, p.PriceMinor, p.Currency, p.Status, p.CreatedAt.AsTime(), p.UpdatedAt.AsTime())
	return err
}

func (s *ProductStore) Find(ctx context.Context, id string) (*productv1.Product, error) {
	return s.scan(s.db.QueryRowContext(ctx, `SELECT id, sku, name, description, price_minor, currency, status, created_at, updated_at FROM products WHERE id=$1`, id))
}

func (s *ProductStore) List(ctx context.Context, status productv1.ProductStatus) ([]*productv1.Product, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, sku, name, description, price_minor, currency, status, created_at, updated_at FROM products WHERE ($1 = 0 OR status = $1) ORDER BY created_at, id`, status)
	if err != nil { return nil, err }
	defer rows.Close()
	var products []*productv1.Product
	for rows.Next() { p, scanErr := s.scan(rows); if scanErr != nil { return nil, scanErr }; products = append(products, p) }
	return products, rows.Err()
}

func (s *ProductStore) Update(ctx context.Context, p *productv1.Product) error {
	_, err := s.db.ExecContext(ctx, `UPDATE products SET sku=$2, name=$3, description=$4, price_minor=$5, currency=$6, status=$7, updated_at=$8 WHERE id=$1`, p.Id, p.Sku, p.Name, p.Description, p.PriceMinor, p.Currency, p.Status, p.UpdatedAt.AsTime())
	return err
}

func (s *ProductStore) Archive(ctx context.Context, id string, status productv1.ProductStatus, updatedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE products SET status=$2, updated_at=$3 WHERE id=$1`, id, status, updatedAt)
	return err
}

type scanner interface{ Scan(...any) error }

func (s *ProductStore) scan(row scanner) (*productv1.Product, error) {
	var p productv1.Product
	var status int32
	var createdAt, updatedAt time.Time
	if err := row.Scan(&p.Id, &p.Sku, &p.Name, &p.Description, &p.PriceMinor, &p.Currency, &status, &createdAt, &updatedAt); err != nil { return nil, err }
	p.Status = productv1.ProductStatus(status)
	p.CreatedAt, p.UpdatedAt = timestamppb.New(createdAt), timestamppb.New(updatedAt)
	return &p, nil
}
