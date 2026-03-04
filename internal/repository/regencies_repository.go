package repository

import (
	"context"
	"database/sql"

	"portal-report-bi/internal/domain"
)

type regencyRepository struct {
	db *sql.DB
}

func NewRegencyRepository(db *sql.DB) domain.RegencyRepository {
	return &regencyRepository{db: db}
}

func (r *regencyRepository) Insert(ctx context.Context, regency *domain.Regency) error {

	query := `
		INSERT INTO regencies (bi_id, regency_name)
		VALUES (:1, :2)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		regency.BIID,
		regency.RegencyName,
	)

	return err
}

func (r *regencyRepository) Get(ctx context.Context) ([]domain.Regency, error) {
	query := `
		SELECT bi_id, regency_name
		FROM regencies
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var regencies []domain.Regency

	for rows.Next() {
		var regency domain.Regency

		err := rows.Scan(
			&regency.BIID,
			&regency.RegencyName,
		)
		if err != nil {
			return nil, err
		}

		regencies = append(regencies, regency)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return regencies, nil
}
func (r *regencyRepository) FindByID(ctx context.Context, id int) (*domain.Regency, error) {

	query := `
		SELECT bi_id, regency_name
		FROM regencies
		WHERE id = :1
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var regency domain.Regency

	err := row.Scan(
		&regency.BIID,
		&regency.RegencyName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrRegencyNotFound
		}
		return nil, err
	}

	return &regency, nil
}

func (r *regencyRepository) Update(ctx context.Context, regency *domain.Regency) (*domain.Regency, error) {
	query := `
		UPDATE regencies
		SET regency_name = :1
		WHERE id = :2
	`
	result, err := r.db.ExecContext(
		ctx,
		query,
		regency.RegencyName,
		regency.ID,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, domain.ErrRegencyNotFound
	}

	return regency, nil
}

func (r *regencyRepository) Delete(ctx context.Context, id int) error {
	query := `
		DELETE FROM regencies
		WHERE id = :1
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return domain.ErrRegencyNotFound
	}

	return nil
}

func (r *regencyRepository) UpdateBIIDByName(ctx context.Context, name string, biid string) error {
	query := `
		UPDATE regencies
		SET bi_id = :1
		WHERE UPPER(TRIM(regency_name)) = UPPER(TRIM(:2))
	`

	_, err := r.db.ExecContext(ctx, query, biid, name)
	return err
}
