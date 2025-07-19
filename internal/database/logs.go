package database

import (
	"ObservabilityServer/internal/model"
	"encoding/json"
)

type LogService interface {
	CreateLog(data model.NewLogData) error
	GetLatestLogsBySessionId(sessionId string, limit, offset int) ([]model.LogEntity, error)
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

func (s *service) GetLatestLogsBySessionId(sessionId string, limit, offset int) ([]model.LogEntity, error) {
	stmt := `
	SELECT id, app_id, session_id, message, data, created_at
	FROM public.ob_logs
	WHERE session_id = $1
	LIMIT $2
	OFFSET $3
	`

	rows, err := s.db.Query(stmt, sessionId, limit, offset)
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
