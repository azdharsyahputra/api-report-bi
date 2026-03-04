package repository

import (
	"context"
	"database/sql"
	"fmt"
	"portal-report-bi/internal/domain"
	"time"
)

type branchCodeBankRepository struct {
	db *sql.DB
}

func NewBranchCodeBankRepository(db *sql.DB) domain.BranchCodeBankRepository {
	return &branchCodeBankRepository{db: db}
}

func (r *branchCodeBankRepository) Insert(ctx context.Context, code *domain.BranchCodeBank) error {
	query := `
		INSERT INTO m_branch_kode_bank(name, branch_code, regencies_code, regencies, office_type, created_at, update_at)
		VALUES(:1, :2, :3, :4, :5, :6, :7)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		code.Name,
		code.BranchCode,
		code.RegenciesCode,
		code.Regencies,
		code.OfficeType,
		time.Now(),
		time.Now(),
	)

	return err
}

func (r *branchCodeBankRepository) GetAll(ctx context.Context, bankName, search string, limit, offset int) ([]domain.BranchCodeBank, int, error) {
	query := `
		SELECT id, 
		       NVL(name, ' ') as name, 
		       NVL(branch_code, ' ') as branch_code, 
		       NVL(regencies_code, ' ') as regencies_code, 
		       NVL(regencies, ' ') as regencies, 
		       NVL(office_type, ' ') as office_type, 
		       NVL(created_at, SYSDATE) as created_at, 
		       NVL(update_at, SYSDATE) as update_at,
		       COUNT(*) OVER() as total_count
		FROM m_branch_kode_bank
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 0

	// Filter spesifik Nama Bank
	if bankName != "" {
		argCount++
		query += fmt.Sprintf(" AND UPPER(name) = UPPER(:%d)", argCount)
		args = append(args, bankName)
	}

	// Search keyword di kolom lain
	if search != "" {
		argCount++
		query += fmt.Sprintf(` 
			AND (
				UPPER(regencies) LIKE UPPER(:%d) OR 
				UPPER(office_type) LIKE UPPER(:%d) OR
				UPPER(name) LIKE UPPER(:%d)
			)`, argCount, argCount+1, argCount+2)

		pattern := "%" + search + "%"
		args = append(args, pattern, pattern, pattern)
		argCount += 3
	}

	// Oracle pagination 12c+ requires ORDER BY to be defined for OFFSET to work reliably
	query += " ORDER BY name ASC, id ASC"

	if limit > 0 {
		query += fmt.Sprintf(" OFFSET :%d ROWS FETCH NEXT :%d ROWS ONLY", argCount+1, argCount+2)
		args = append(args, offset, limit)
		argCount += 2
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []domain.BranchCodeBank
	var totalCount int
	for rows.Next() {
		var b domain.BranchCodeBank
		err := rows.Scan(
			&b.ID, &b.Name, &b.BranchCode, &b.RegenciesCode, &b.Regencies, &b.OfficeType, &b.CreatedAt, &b.UpdateAt,
			&totalCount,
		)
		if err != nil {
			return nil, 0, err
		}
		results = append(results, b)
	}

	return results, totalCount, nil
}

func (r *branchCodeBankRepository) FindByID(ctx context.Context, id int) (*domain.BranchCodeBank, error) {
	query := `
		SELECT id, name, branch_code, regencies_code, regencies, office_type, created_at, update_at
		FROM m_branch_kode_bank
		WHERE id = :1
	`
	row := r.db.QueryRowContext(ctx, query, id)
	var b domain.BranchCodeBank
	err := row.Scan(&b.ID, &b.Name, &b.BranchCode, &b.RegenciesCode, &b.Regencies, &b.OfficeType, &b.CreatedAt, &b.UpdateAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *branchCodeBankRepository) Update(ctx context.Context, code *domain.BranchCodeBank) (*domain.BranchCodeBank, error) {
	query := `
		UPDATE m_branch_kode_bank
		SET name = :1, branch_code = :2, regencies_code = :3, regencies = :4, office_type = :5, update_at = :6
		WHERE id = :7
	`
	res, err := r.db.ExecContext(ctx, query, code.Name, code.BranchCode, code.RegenciesCode, code.Regencies, code.OfficeType, time.Now(), code.ID)
	if err != nil {
		return nil, err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return nil, domain.ErrInvalidId
	}

	return code, nil
}

func (r *branchCodeBankRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM m_branch_kode_bank WHERE id = :1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	count, _ := res.RowsAffected()
	if count == 0 {
		return domain.ErrInvalidId
	}
	return nil
}

func (r *branchCodeBankRepository) BulkInsert(ctx context.Context, codes []domain.BranchCodeBank) error {
	if len(codes) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	batchSize := 500
	for i := 0; i < len(codes); i += batchSize {
		end := i + batchSize
		if end > len(codes) {
			end = len(codes)
		}
		batch := codes[i:end]

		if err := r.executeBatchInsert(ctx, tx, batch); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *branchCodeBankRepository) executeBatchInsert(ctx context.Context, tx *sql.Tx, batch []domain.BranchCodeBank) error {
	for _, code := range batch {

		query := `
			MERGE INTO m_branch_kode_bank t
			USING (SELECT :1 as name, :2 as branch_code FROM dual) s
			ON (t.name = s.name AND t.branch_code = s.branch_code)
			WHEN MATCHED THEN
				UPDATE SET t.regencies_code = :3, t.regencies = :4, t.office_type = :5, t.update_at = :6
			WHEN NOT MATCHED THEN
				INSERT (name, branch_code, regencies_code, regencies, office_type, created_at, update_at)
				VALUES (:7, :8, :9, :10, :11, :12, :13)
		`
		now := time.Now()
		_, err := tx.ExecContext(ctx, query,
			code.Name, code.BranchCode,
			code.RegenciesCode, code.Regencies, code.OfficeType, now,
			code.Name, code.BranchCode, code.RegenciesCode, code.Regencies, code.OfficeType, now, now,
		)
		if err != nil {

			continue
		}
	}
	return nil
}
