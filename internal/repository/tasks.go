package db

import (
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"golang.org/x/net/context"
)

func (q *Queries) GetTaskListByFilter(ctx context.Context, filter TaskListFilter) ([]Task, error) {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
	query := builder.Select("*").From("tasks").Limit(uint64(filter.Limit)).Offset(uint64(filter.Offset))

	if filter.Difficulties != nil {
		query = query.Where(sq.Eq{"difficulty": filter.Difficulties})
	}
	if filter.SortBy != nil {
		query = query.OrderBy(*filter.SortBy)
	}
	if filter.TopicIDs != nil {
		query = query.Where("topics @> ?", filter.TopicIDs)
	}
	if len(filter.Title) > 0 {
		query = query.Where("title LIKE '%' || ? || '%'", filter.Title)
	}
	sql, args, err := query.ToSql()
	fmt.Println(sql, args)
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
