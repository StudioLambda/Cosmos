package contract_test

import (
	"context"
	"errors"
	"testing"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/studiolambda/cosmos/contract"
	"github.com/studiolambda/cosmos/contract/mock"
)

type rowsMock struct {
	nextResults []bool
	scanFuncs   []func(dest any) error
	err         error
	closeErr    error
	closeCalled bool
	nextIndex   int
	scanIndex   int
}

func (rows *rowsMock) Next() bool {
	if rows.nextIndex >= len(rows.nextResults) {
		return false
	}

	result := rows.nextResults[rows.nextIndex]
	rows.nextIndex++

	return result
}

func (rows *rowsMock) Scan(dest any) error {
	if rows.scanIndex >= len(rows.scanFuncs) {
		return nil
	}

	f := rows.scanFuncs[rows.scanIndex]
	rows.scanIndex++

	return f(dest)
}

func (rows *rowsMock) Err() error {
	return rows.err
}

func (rows *rowsMock) Close() error {
	rows.closeCalled = true

	return rows.closeErr
}

func TestErrDatabaseNoRowsMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"no database rows were found",
		contract.ErrDatabaseNoRows.Error(),
	)
}

func TestErrDatabaseNoRowsIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrDatabaseNoRows)
}

func TestErrDatabaseNestedTransactionMessage(t *testing.T) {
	t.Parallel()

	require.Equal(
		t,
		"nested transactions are not supported",
		contract.ErrDatabaseNestedTransaction.Error(),
	)
}

func TestErrDatabaseNestedTransactionIsNonNil(t *testing.T) {
	t.Parallel()

	require.NotNil(t, contract.ErrDatabaseNestedTransaction)
}

func TestDatabaseSelectReturnsCollectionSlice(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Select(context.Background(), "SELECT id FROM users", testifymock.Anything, []any{1}).RunAndReturn(
		func(ctx context.Context, query string, dest any, args ...any) error {
			users, ok := dest.(*[]user)
			require.True(t, ok)
			require.Equal(t, []any{1}, args)

			*users = []user{{ID: 1}, {ID: 2}}

			return nil
		},
	)

	database := contract.NewDatabase(driver)

	users, err := database.Select[user](context.Background(), "SELECT id FROM users", 1)

	require.NoError(t, err)
	require.Equal(t, []user{{ID: 1}, {ID: 2}}, users.Items())
}

func TestDatabaseSelectReturnsDriverError(t *testing.T) {
	t.Parallel()

	expected := errors.New("select failed")
	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Select(context.Background(), "SELECT id FROM users", testifymock.Anything).Return(expected)

	database := contract.NewDatabase(driver)

	users, err := database.Select[int](context.Background(), "SELECT id FROM users")

	require.ErrorIs(t, err, expected)
	require.True(t, users.IsEmpty())
}

func TestDatabaseSelectNamedReturnsCollectionSlice(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	arg := map[string]any{"account_id": 42}
	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().SelectNamed(context.Background(), "SELECT id FROM users WHERE account_id=:account_id", testifymock.Anything, arg).RunAndReturn(
		func(ctx context.Context, query string, dest any, namedArg any) error {
			users, ok := dest.(*[]user)
			require.True(t, ok)
			require.Equal(t, arg, namedArg)

			*users = []user{{ID: 3}, {ID: 4}}

			return nil
		},
	)

	database := contract.NewDatabase(driver)

	users, err := database.SelectNamed[user](context.Background(), "SELECT id FROM users WHERE account_id=:account_id", arg)

	require.NoError(t, err)
	require.Equal(t, []user{{ID: 3}, {ID: 4}}, users.Items())
}

func TestDatabaseSelectNamedReturnsDriverError(t *testing.T) {
	t.Parallel()

	expected := errors.New("select named failed")
	arg := map[string]any{"account_id": 42}
	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().SelectNamed(context.Background(), "SELECT id FROM users WHERE account_id=:account_id", testifymock.Anything, arg).Return(expected)

	database := contract.NewDatabase(driver)

	users, err := database.SelectNamed[int](context.Background(), "SELECT id FROM users WHERE account_id=:account_id", arg)

	require.ErrorIs(t, err, expected)
	require.True(t, users.IsEmpty())
}

func TestDatabaseCursorYieldsRowsLazily(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	rows := &rowsMock{
		nextResults: []bool{true, true, false},
		scanFuncs: []func(dest any) error{
			func(dest any) error {
				user, ok := dest.(*user)
				require.True(t, ok)
				user.ID = 1

				return nil
			},
			func(dest any) error {
				user, ok := dest.(*user)
				require.True(t, ok)
				user.ID = 2

				return nil
			},
		},
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Query(context.Background(), "SELECT id FROM users", []any{1}).Return(rows, nil)

	database := contract.NewDatabase(driver)

	items, err := database.Cursor[user](context.Background(), "SELECT id FROM users", 1).Items()

	require.NoError(t, err)
	require.Equal(t, []user{{ID: 1}, {ID: 2}}, items)
	require.True(t, rows.closeCalled)
}

func TestDatabaseCursorReturnsQueryError(t *testing.T) {
	t.Parallel()

	expected := errors.New("query failed")
	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Query(context.Background(), "SELECT id FROM users").Return(nil, expected)

	database := contract.NewDatabase(driver)

	items, err := database.Cursor[int](context.Background(), "SELECT id FROM users").Items()

	require.ErrorIs(t, err, expected)
	require.Empty(t, items)
}

func TestDatabaseCursorStopsOnScanErrorAndClosesRows(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	expected := errors.New("scan failed")
	rows := &rowsMock{
		nextResults: []bool{true, true},
		scanFuncs: []func(dest any) error{
			func(dest any) error {
				user := dest.(*user)
				user.ID = 1

				return nil
			},
			func(dest any) error {
				return expected
			},
		},
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Query(context.Background(), "SELECT id FROM users").Return(rows, nil)

	database := contract.NewDatabase(driver)

	items, err := database.Cursor[user](context.Background(), "SELECT id FROM users").Items()

	require.Equal(t, []user{{ID: 1}}, items)
	require.ErrorIs(t, err, expected)
	require.True(t, rows.closeCalled)
}

func TestDatabaseCursorReturnsIterationError(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	expected := errors.New("iteration failed")
	rows := &rowsMock{
		nextResults: []bool{true, false},
		scanFuncs: []func(dest any) error{
			func(dest any) error {
				user := dest.(*user)
				user.ID = 1

				return nil
			},
		},
		err: expected,
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Query(context.Background(), "SELECT id FROM users").Return(rows, nil)

	database := contract.NewDatabase(driver)

	items, err := database.Cursor[user](context.Background(), "SELECT id FROM users").Items()

	require.Equal(t, []user{{ID: 1}}, items)
	require.ErrorIs(t, err, expected)
	require.True(t, rows.closeCalled)
}

func TestDatabaseCursorStopsEarlyAndClosesRows(t *testing.T) {
	t.Parallel()

	rows := &rowsMock{
		nextResults: []bool{true, true, true},
		scanFuncs: []func(dest any) error{
			func(dest any) error {
				value := dest.(*int)
				*value = 1

				return nil
			},
			func(dest any) error {
				value := dest.(*int)
				*value = 2

				return nil
			},
			func(dest any) error {
				value := dest.(*int)
				*value = 3

				return nil
			},
		},
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().Query(context.Background(), "SELECT id FROM users").Return(rows, nil)

	database := contract.NewDatabase(driver)

	err := database.Cursor[int](context.Background(), "SELECT id FROM users").Each(func(i int, v int) error {
		if v == 2 {
			return errors.New("stop")
		}

		return nil
	})

	require.Error(t, err)
	require.True(t, rows.closeCalled)
}

func TestDatabaseCursorNamedYieldsRowsLazily(t *testing.T) {
	t.Parallel()

	type user struct {
		ID int
	}

	arg := map[string]any{"account_id": 42}
	rows := &rowsMock{
		nextResults: []bool{true, true, false},
		scanFuncs: []func(dest any) error{
			func(dest any) error {
				user := dest.(*user)
				user.ID = 5

				return nil
			},
			func(dest any) error {
				user := dest.(*user)
				user.ID = 6

				return nil
			},
		},
	}

	driver := mock.NewDatabaseDriverMock(t)
	driver.EXPECT().QueryNamed(context.Background(), "SELECT id FROM users WHERE account_id=:account_id", arg).Return(rows, nil)

	database := contract.NewDatabase(driver)

	items, err := database.CursorNamed[user](context.Background(), "SELECT id FROM users WHERE account_id=:account_id", arg).Items()

	require.NoError(t, err)
	require.Equal(t, []user{{ID: 5}, {ID: 6}}, items)
	require.True(t, rows.closeCalled)
}
