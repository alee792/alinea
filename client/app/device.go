package main

import "database/sql"

// Device describes a Alinea server on a Device
type Device struct {
	ID          string             `json:"id,omitempty"`
	Name        string             `json:"name"`
	URL         string             `json:"url"`
	Location    string             `json:"location"`
	Active      ConvertibleBoolean `json:"active"`
	Description string             `json:"description"`
}

// DeviceData is passed to templates
type DeviceData struct {
	PageTitle string
	Actions   map[string]string
	Devices   []Device
}

func (d *Device) create(db *sql.DB) (err error) {
	return db.QueryRow(
		"INSERT INTO devices(name, url, active, description) VALUES($1, $2, $3, $4) RETURNING id",
		d.Name, d.URL, d.Active, d.Description,
	).Scan(&d.ID)
}

func (d *Device) getDevice(db *sql.DB, id string) (err error) {
	return db.QueryRow(
		"SELECT id, name, url FROM devices WHERE id = $1",
		id,
	).Scan(&d.ID, &d.Name, &d.URL)
}

func getAll(db *sql.DB, limit, offset int) (ds []Device, err error) {
	rows, err := db.Query(
		"SELECT id, name, location, active, description FROM devices  ORDER BY name ASC LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var d Device
		var location sql.NullString
		var description sql.NullString
		err = rows.Scan(&d.ID, &d.Name, &location, &d.Active, &description)
		if err != nil {
			return
		}
		if location.Valid {
			d.Location = location.String
		} else {
			d.Location = ""
		}
		if description.Valid {
			d.Description = description.String
		} else {
			d.Location = ""
		}
		ds = append(ds, d)
	}
	return
}
