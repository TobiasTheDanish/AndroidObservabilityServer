package database

import (
	"ObservabilityServer/internal/model"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
)

type LogService interface {
	CreateLog(data model.NewLogData) error
	GetLatestLogsBySessionId(sessionId string, limit, offset int, filters []LogFilter) ([]model.LogEntity, error)
}

type LogFilter struct {
	Key      string
	Value    string
	Operator string
}

func (s *service) CreateLog(data model.NewLogData) error {
	stmt := `
	INSERT INTO public.ob_logs 
	(app_id, session_id, message, data, created_at)
	VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.db.Exec(
		stmt,
		data.AppId,
		data.SessionId,
		data.Message,
		jsonBuildObjectContent(data.Data),
		data.CreatedAt,
	)

	return err
}

func (s *service) GetLatestLogsBySessionId(sessionId string, limit, offset int, filters []LogFilter) ([]model.LogEntity, error) {
	if filters == nil {
		filters = make([]LogFilter, 0)
	}
	filters = slices.Insert(filters, 0, LogFilter{
		Key:      "session_id",
		Operator: "eq",
		Value:    sessionId,
	})
	where, whereValues := logFilterToWhereClause(filters, 3)

	stmt := `
	SELECT id, app_id, session_id, message, data, created_at
	FROM public.ob_logs
	WHERE ` + where + `
	ORDER BY created_at DESC
	LIMIT $1
	OFFSET $2
	`

	values := make([]any, 0)
	values = append(values, limit, offset)
	for _, wVal := range whereValues {
		values = append(values, wVal)
	}

	rows, err := s.db.Query(stmt, values...)
	if err != nil {
		return nil, err
	}

	entities := make([]model.LogEntity, 0)
	for rows.Next() {
		var entityData []byte
		var e model.LogEntity
		err = rows.Scan(
			&e.Id,
			&e.AppId,
			&e.SessionId,
			&e.Message,
			&entityData,
			&e.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(entityData, &e.Data)
		if err != nil {
			return nil, err
		}

		entities = append(entities, e)
	}

	return entities, nil
}

func logFilterToWhereClause(filters []LogFilter, startPlaceholder int) (string, []string) {
	if filters == nil {
		return "", []string{}
	}

	filterCount := len(filters)
	values := make([]string, filterCount, filterCount)
	var sb strings.Builder

	for i, filter := range filters {
		sb.WriteString(filter.Key)

		switch filter.Operator {
		case "eq":
			{
				sb.WriteString(" = ")
				values[i] = filter.Value
			}
		case "greater":
			{
				sb.WriteString(" > ")
				values[i] = filter.Value
			}
		case "less":
			{
				sb.WriteString(" < ")
				values[i] = filter.Value
			}
		case "contains":
			{
				sb.WriteString(" LIKE ")
				values[i] = "%" + filter.Value + "%"
			}
		case "starts":
			{
				sb.WriteString(" LIKE ")
				values[i] = filter.Value + "%"
			}
		case "ends":
			{
				sb.WriteString(" LIKE ")
				values[i] = "%" + filter.Value
			}

		default:
			continue
		}

		sb.WriteString("$")
		sb.WriteString(strconv.Itoa(startPlaceholder + i))

		if i < filterCount-1 {
			sb.WriteString("\nAND ")
		}
	}

	return sb.String(), values
}
