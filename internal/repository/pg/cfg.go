package pg

import (
	"encoding/json"
	"fmt"
	"myapp/internal/entity"
)

func (s *Database) EditCfgVal(id, val string) error {
	q := `UPDATE cfg SET val = $1 WHERE id = $2`
	_, err := s.Exec(q, val, id)
	if err != nil {
		return fmt.Errorf("EditCfgVal err: %w", err)
	}
	return nil
}

func (s *Database) GetCfgValById(id string) (entity.Cfg, error) {
	q := `
		SELECT coalesce((
			SELECT to_json(c)
	  		FROM cfg as c
	  		WHERE id = $1
		), '{}'::json)
	`
	var u entity.Cfg
	var data []byte
	err := s.QueryRow(q, id).Scan(&data)
	if err != nil {
		return u, fmt.Errorf("GetCfgValById Scan: %v", err)
	}
	if err := json.Unmarshal(data, &u); err != nil {
		return u, fmt.Errorf("GetCfgValById Unmarshal: %v", err)
	}
	return u, nil
}
