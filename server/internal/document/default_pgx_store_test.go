package document_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"server/internal/document"
	"server/internal/pglib"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

var errNoRows = errors.New("no rows in result set")
var errQueryFailed = errors.New("query failed")
var errIterFailed = errors.New("iteration failed")

func makeRowsMock(numRows int) *pglib.RowsMock {
	rows := &pglib.RowsMock{}
	callCount := 0
	rows.CloseFunc = func() {}
	rows.CommandTagFunc = func() pgconn.CommandTag { return pgconn.CommandTag{} }
	rows.ConnFunc = func() *pgx.Conn { return nil }
	rows.NextFunc = func() bool {
		callCount++
		return callCount <= numRows
	}
	rows.ErrFunc = func() error { return nil }
	rows.ScanFunc = func(dest ...any) error {
		for i, d := range dest {
			switch d.(type) {
			case *pgtype.UUID:
				u := d.(*pgtype.UUID)
				u.Valid = true
				u.Bytes = uuid.New()
			case *string:
				s := d.(*string)
				switch i {
				case 2:
					*s = "passport.pdf"
				case 3:
					*s = "/files/123"
				case 4:
					*s = "application/pdf"
				case 5:
					*s = "abc123"
				case 6:
					*s = "pending"
				}
			case *int:
				*d.(*int) = 0
			case *time.Time:
				*d.(*time.Time) = time.Now()
			}
		}
		return nil
	}
	rows.FieldDescriptionsFunc = func() []pgconn.FieldDescription { return nil }
	rows.RawValuesFunc = func() [][]byte { return nil }
	rows.ValuesFunc = func() ([]any, error) { return nil, nil }
	return rows
}

func TestDocumentStore_Create(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*pglib.PoolMock)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						for i, d := range dest {
							switch d.(type) {
							case *pgtype.UUID:
								u := d.(*pgtype.UUID)
								u.Valid = true
								u.Bytes = uuid.New()
							case *string:
								s := d.(*string)
								switch i {
								case 2:
									*s = "passport.pdf"
								case 3:
									*s = "/files/123"
								case 4:
									*s = "application/pdf"
								case 6:
									*s = "pending"
								}
							case *int:
								*d.(*int) = 0
							case *time.Time:
								*d.(*time.Time) = time.Now()
							}
						}
						return nil
					}
					return row
				}
			},
		},
		{
			name:    "query error",
			wantErr: true,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						return errQueryFailed
					}
					return row
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := &pglib.PoolMock{}
			tt.setupMock(mockPool)

			store := document.NewDefaultPgxStore(mockPool)
			result, err := store.Create(context.Background(), uuid.New(), uuid.New(), "passport.pdf", "/files/123", "application/pdf")

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestDocumentStore_GetByID(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*pglib.PoolMock)
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						for i, d := range dest {
							switch d.(type) {
							case *pgtype.UUID:
								u := d.(*pgtype.UUID)
								u.Valid = true
								u.Bytes = uuid.New()
							case *string:
								s := d.(*string)
								switch i {
								case 2:
									*s = "passport.pdf"
								case 3:
									*s = "/files/123"
								case 4:
									*s = "application/pdf"
								case 5:
									*s = "abc123"
								case 6:
									*s = "pending"
								}
							case *int:
								*d.(*int) = 0
							case *time.Time:
								*d.(*time.Time) = time.Now()
							}
						}
						return nil
					}
					return row
				}
			},
		},
		{
			name:    "not found",
			wantErr: true,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						return errNoRows
					}
					return row
				}
			},
		},
		{
			name:    "query error",
			wantErr: true,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						return errQueryFailed
					}
					return row
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := &pglib.PoolMock{}
			tt.setupMock(mockPool)

			store := document.NewDefaultPgxStore(mockPool)
			result, err := store.GetByID(context.Background(), uuid.New())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestDocumentStore_GetByUserID(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*pglib.PoolMock)
		wantCount int
		wantErr   bool
	}{
		{
			name:      "success multiple",
			wantCount: 2,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryFunc = func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
					return makeRowsMock(2), nil
				}
			},
		},
		{
			name:      "success empty",
			wantCount: 0,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryFunc = func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
					return makeRowsMock(0), nil
				}
			},
		},
		{
			name:    "query error",
			wantErr: true,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryFunc = func(ctx context.Context, query string, args ...any) (pgx.Rows, error) {
					return nil, errQueryFailed
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := &pglib.PoolMock{}
			tt.setupMock(mockPool)

			store := document.NewDefaultPgxStore(mockPool)
			results, err := store.GetByUserID(context.Background(), uuid.New())

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, results)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)
			}
		})
	}
}

func TestDocumentStore_Update(t *testing.T) {
	tests := []struct {
		name      string
		req       *document.UpdateRequest
		wantErr   bool
		setupMock func(*pglib.PoolMock)
	}{
		{
			name: "update status",
			req:  &document.UpdateRequest{Status: "verified"},
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						for i, d := range dest {
							switch d.(type) {
							case *pgtype.UUID:
								u := d.(*pgtype.UUID)
								u.Valid = true
								u.Bytes = uuid.New()
							case *string:
								s := d.(*string)
								switch i {
								case 2:
									*s = "passport.pdf"
								case 3:
									*s = "/files/123"
								case 4:
									*s = "application/pdf"
								case 5:
									*s = "abc123"
								case 6:
									*s = "verified"
								}
							case *int:
								*d.(*int) = 0
							case *time.Time:
								*d.(*time.Time) = time.Now()
							}
						}
						return nil
					}
					return row
				}
			},
		},
		{
			name: "update checksum",
			req:  &document.UpdateRequest{Checksum: "new-checksum"},
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						for i, d := range dest {
							switch d.(type) {
							case *pgtype.UUID:
								u := d.(*pgtype.UUID)
								u.Valid = true
								u.Bytes = uuid.New()
							case *string:
								s := d.(*string)
								switch i {
								case 2:
									*s = "passport.pdf"
								case 3:
									*s = "/files/123"
								case 4:
									*s = "application/pdf"
								case 5:
									*s = "new-checksum"
								case 6:
									*s = "pending"
								}
							case *int:
								*d.(*int) = 0
							case *time.Time:
								*d.(*time.Time) = time.Now()
							}
						}
						return nil
					}
					return row
				}
			},
		},
		{
			name: "update both",
			req:  &document.UpdateRequest{Status: "verified", Checksum: "new-checksum"},
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						for i, d := range dest {
							switch d.(type) {
							case *pgtype.UUID:
								u := d.(*pgtype.UUID)
								u.Valid = true
								u.Bytes = uuid.New()
							case *string:
								s := d.(*string)
								switch i {
								case 2:
									*s = "passport.pdf"
								case 3:
									*s = "/files/123"
								case 4:
									*s = "application/pdf"
								case 5:
									*s = "new-checksum"
								case 6:
									*s = "verified"
								}
							case *int:
								*d.(*int) = 0
							case *time.Time:
								*d.(*time.Time) = time.Now()
							}
						}
						return nil
					}
					return row
				}
			},
		},
		{
			name:    "query error",
			req:     &document.UpdateRequest{Status: "verified"},
			wantErr: true,
			setupMock: func(m *pglib.PoolMock) {
				m.QueryRowFunc = func(ctx context.Context, query string, args ...any) pgx.Row {
					row := &pglib.RowMock{}
					row.ScanFunc = func(dest ...any) error {
						return errQueryFailed
					}
					return row
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPool := &pglib.PoolMock{}
			tt.setupMock(mockPool)

			store := document.NewDefaultPgxStore(mockPool)
			result, err := store.Update(context.Background(), uuid.New(), tt.req)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
