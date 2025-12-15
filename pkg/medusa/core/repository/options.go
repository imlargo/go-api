package repository

type RepositoryOption func(*Repository)

func WithEntity(entity any) RepositoryOption {
	return func(r *Repository) {
		r.entity = entity
	}
}
