package plugins

import (
	"archive/zip"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/database"
)

func ExportObject(c *rest.Context, o *database.Object, opath string, zipWriter *zip.Writer) error {
	f, err := zipWriter.Create(filepath.Join(opath, "object.json"))
	if err != nil {
		return err
	}
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	data, err := c.Request(c, "GET", "/api/objects/"+o.ID+"/data", nil, nil)
	if err != nil {
		return err
	}
	defer data.Close()

	f, err = zipWriter.Create(filepath.Join(opath, "data"))
	if err != nil {
		return err
	}

	_, err = io.Copy(f, data)

	return err
}

type ExportAppOptions struct {
	IncludeObjects bool
}

func ExportApp(c *rest.Context, app *database.App, opath string, zipWriter *zip.Writer, opt *ExportAppOptions) error {
	f, err := zipWriter.Create(filepath.Join(opath, "app.json"))
	if err != nil {
		return err
	}
	b, err := json.Marshal(app)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	if opt != nil && opt.IncludeObjects {
		objs, err := c.DB.ListObjects(&database.ListObjectsOptions{
			App:               &app.ID,
			ReadObjectOptions: database.ReadObjectOptions{Icon: true},
		})
		if err != nil {
			return err
		}
		for _, o := range objs {
			err = ExportObject(c, o, filepath.Join(opath, "objects", o.ID), zipWriter)
			if err != nil {
				c.Log.Warn("Failed to export object", o.ID, ":", err)
			}
		}
	}

	return nil
}

type ExportUserOptions struct {
	IncludeApps bool
}

func ExportUser(c *rest.Context, u *database.User, opath string, zipWriter *zip.Writer, opt *ExportUserOptions) error {
	f, err := zipWriter.Create(filepath.Join(opath, "user.json"))
	if err != nil {
		return err
	}
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	f, err = zipWriter.Create(filepath.Join(opath, "settings.json"))
	if err != nil {
		return err
	}
	settings, err := c.DB.ReadUserSettings(*u.UserName)
	if err != nil {
		return err
	}
	b, err = json.Marshal(settings)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	if opt != nil && opt.IncludeApps {
		apps, err := c.DB.ListApps(&database.ListAppOptions{
			Owner:          u.UserName,
			ReadAppOptions: database.ReadAppOptions{Icon: true},
		})
		if err != nil {
			return err
		}
		for _, a := range apps {
			err = ExportApp(c, a, filepath.Join(opath, "apps", a.ID), zipWriter, nil)
			if err != nil {
				return err
			}
		}
	}

	objs, err := c.DB.ListObjects(&database.ListObjectsOptions{
		Owner:             u.UserName,
		ReadObjectOptions: database.ReadObjectOptions{Icon: true},
	})
	if err != nil {
		return err
	}
	for _, o := range objs {
		err = ExportObject(c, o, filepath.Join(opath, "objects", o.ID), zipWriter)
		if err != nil {
			c.Log.Warn("Failed to export object", o.ID, ":", err)
		}
	}

	return nil
}
