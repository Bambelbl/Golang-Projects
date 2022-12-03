package user

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestAuthorize(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	testUser := &User{
		ID:       RandStringRunes(),
		Username: "mem",
		password: "kek12345678",
	}

	// NewRepo
	repo := NewMySQLRepo(db)
	err = repo.DB.Ping()
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}

	// NoUserError
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnError(sql.ErrNoRows)

	_, err = repo.Authorize(testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	// SELECT Error
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnError(errors.New("UserID exist"))

	_, err = repo.Authorize(testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// BadPassErr
	rows := sqlmock.NewRows([]string{"id", "username", "pass"})
	rows = rows.AddRow(testUser.ID, testUser.Username, "nekek")
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnRows(rows)

	_, err = repo.Authorize(testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// ok query
	rows = sqlmock.NewRows([]string{"id", "username", "pass"})
	rows = rows.AddRow(testUser.ID, testUser.Username, testUser.password)
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnRows(rows)

	user, err := repo.Authorize(testUser.Username, testUser.password)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
	if !reflect.DeepEqual(user, testUser) {
		t.Errorf("results not match, want %v, have %v", testUser, user)
		return
	}
}

func TestCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	testUser := &User{
		ID:       RandStringRunes(),
		Username: "mem",
		password: "kek12345678",
	}

	// NewRepo
	repo := NewMySQLRepo(db)
	err = repo.DB.Ping()
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}

	// ok query
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(testUser.ID, testUser.Username, testUser.password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = repo.AddUser(testUser.ID, testUser.Username, testUser.password)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// internal error in INSERT
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnError(sql.ErrNoRows)
	mock.
		ExpectExec("INSERT INTO users").
		WithArgs(testUser.ID, testUser.Username, testUser.password).
		WillReturnError(errors.New("INSERT err"))

	err = repo.AddUser(testUser.ID, testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// UserExist err
	rows := sqlmock.NewRows([]string{"id", "username", "pass"})
	userExpect := &User{
		ID:       RandStringRunes(),
		Username: "mem",
		password: "kek12345678",
	}
	rows = rows.AddRow(userExpect.ID, userExpect.Username, userExpect.password)
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnRows(rows)

	err = repo.AddUser(testUser.ID, testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err = mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}

	// Internal error
	mock.ExpectQuery("SELECT id, username, pass FROM users WHERE").
		WithArgs(testUser.Username).
		WillReturnError(errors.New("UserID exist"))

	err = repo.AddUser(testUser.ID, testUser.Username, testUser.password)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
