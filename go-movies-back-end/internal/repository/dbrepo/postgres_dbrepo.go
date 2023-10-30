package dbrepo

import (
	"backend/internal/models"
	"context"
	"database/sql"
	"time"
)

type PostgresDBRepo struct {
	DB *sql.DB
}

const dbTimout = time.Second * 3

func (m *PostgresDBRepo) Connection() *sql.DB {
	return m.DB
}

func (m *PostgresDBRepo) AllMovies() ([]*models.Movie, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimout)
	defer cancel()

	// In Golang, you cannot do anything with the database null value. So,
	// we use coalesce. coalesce(image, '') => it does => if there is value in image
	// field it reurns that value if there is no value or null then it return empty
	// space('') in null's place. We're using coalesce in image column but not in any
	// other column because there is a strong chance that we wouldn't be able to provide
	// image for every movie.
	query := `
		select
			id, title, release_date, runtime,
			mpaa_rating, description, coalesce(image, ''),
			created_at, updated_at
		from
			movies
		order by
			title
	`

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var movies []*models.Movie

	// Go through each of the rows one at a time
	for rows.Next() {
		var movie models.Movie
		// scan the current row into movie variable
		err := rows.Scan(
			&movie.ID,
			&movie.Title,
			&movie.ReleaseDate,
			&movie.RunTime,
			&movie.MPAARating,
			&movie.Description,
			&movie.Image,
			&movie.CreatedAt,
			&movie.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		movies = append(movies, &movie)
	}

	return movies, nil
}

func (m *PostgresDBRepo) GetUserByEmail(email string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimout)
	defer cancel()

	query := `select id, email, first_name, last_name, password,
			created_at, updated_at from users where email = $1`

	var user models.User
	// QueryRowContext => we're making a query that returns at most one row. And email is the
	// substitution of parameter $1 in query.
	row := m.DB.QueryRowContext(ctx, query, email)

	// And now, we have a row
	err := row.Scan(
		&user.ID,        // we're going to read first return value into user.ID
		&user.Email,     // we're going to read second return value into user.Email
		&user.FirstName, // we're going to read third return value into user.FirstName
		&user.LastName,  // we're going to read fourth return value into user.LastName
		&user.Password,  // we're going to read fifth return value into user.Password
		&user.CreatedAt, // we're going to read sixth return value into user.CreatedAt
		&user.UpdatedAt, // we're going to read seventh return value into user.UpdatedAt
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (m *PostgresDBRepo) GetUserByID(id int) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimout)
	defer cancel()

	query := `select id, email, first_name, last_name, password,
			created_at, updated_at from users where id = $1`

	var user models.User
	// QueryRowContext => we're making a query that returns at most one row. And email is the
	// substitution of parameter $1 in query.
	row := m.DB.QueryRowContext(ctx, query, id)

	// And now, we have a row
	err := row.Scan(
		&user.ID,        // we're going to read first return value into user.ID
		&user.Email,     // we're going to read second return value into user.Email
		&user.FirstName, // we're going to read third return value into user.FirstName
		&user.LastName,  // we're going to read fourth return value into user.LastName
		&user.Password,  // we're going to read fifth return value into user.Password
		&user.CreatedAt, // we're going to read sixth return value into user.CreatedAt
		&user.UpdatedAt, // we're going to read seventh return value into user.UpdatedAt
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
