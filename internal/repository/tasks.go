package db

import (
	sq "github.com/Masterminds/squirrel"
	"golang.org/x/net/context"
)

var taskSortColumns = map[string]string{
	"name":       "title",
	"difficulty": "difficulty",
	"score":      "score",
}

func (q *Queries) GetTaskListByFilter(ctx context.Context, filter TaskListFilter) ([]Task, error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query := builder.Select("*").From("tasks").Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset))

	if filter.Difficulties != nil {
		query = query.Where(sq.Eq{"difficulty": filter.Difficulties})
	}
	if filter.SortBy != nil {
		if column, ok := taskSortColumns[*filter.SortBy]; ok {
			query = query.OrderBy(column)
		}
	}
	if filter.TopicIDs != nil {
		query = query.Where("topics @> ?", filter.TopicIDs)
	}
	if len(filter.Title) > 0 {
		query = query.Where("title LIKE '%' || ? || '%'", filter.Title)
	}
	if filter.CourseID != nil {
		query = query.Where(sq.Eq{"course_id": *filter.CourseID})
	}
	if filter.IsPublic != nil {
		query = query.Where(sq.Eq{"is_public": *filter.IsPublic})
	}
	if filter.UserID != nil {
		query = query.Where(sq.Eq{"user_id": *filter.UserID})
	}
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := q.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Task
	for rows.Next() {
		var i Task
		if err := rows.Scan(
			&i.ID,
			&i.UserID,
			&i.Title,
			&i.Description,
			&i.Difficulty,
			&i.Passes,
			&i.Score,
			&i.Topics,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.CourseID,
			&i.IsPublic,
			&i.VerificationFile,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
