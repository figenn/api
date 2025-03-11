package auth

import (
	"figenn/internal/database"
	"figenn/internal/user"

	"github.com/bluele/gcache"
)

type Repository struct {
	s     database.Service
	cache gcache.Cache
}

func NewRepository(db database.Service) *Repository {
	return &Repository{
		s:     db,
		cache: gcache.New(20).LRU().Build(),
	}
}

func (r *Repository) UserExistsByEmail(email string) (bool, error) {
	var count int
	err := r.s.DB().QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (r *Repository) CreateUser(user *user.User) error {
	query := `INSERT INTO users (email, password, first_name,last_name) VALUES ($1, $2, $3, $4) RETURNING id`
	return r.s.DB().QueryRow(query, user.Email, user.Password, user.FirstName, user.LastName).Scan(&user.ID)
}

func (r *Repository) GetUserByEmail(email string) (*user.User, error) {
	var u user.User
	err := r.s.DB().QueryRow("SELECT id, email, first_name, last_name, password FROM users WHERE email = $1", email).Scan(&u.ID, &u.Email, &u.FirstName, &u.LastName, &u.Password)
	if err != nil {
		return nil, err
	}

	return &u, nil
}
