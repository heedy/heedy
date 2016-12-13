package commands

import (
	"config"
	"connectordb"
	"connectordb/datastream"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"util"
	"util/datapoint"

	log "github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

// ExportInfo contains the information necessary for an importer to import the database
type ExportInfo struct {
	Version     int    // The export format version
	ConnectorDB string // The version of ConnectorDB that generated the export
}

//WriteStreamDataToFile writes the given DataRange to a file as a json array
func WriteStreamDataToFile(filename string, dr datastream.DataRange) error {
	jreader, err := datapoint.NewJsonArrayReader(dr)
	if err == io.EOF {
		// There is no data in the stream
		return ioutil.WriteFile(filename, []byte("[]"), 0666)
	}
	if err != nil {
		return err
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, jreader)
	return err

}

// ExportCmd generates a data dump which can later be imported
var ExportCmd = &cobra.Command{
	Use:   "export [config file path or database directory] [export directory]",
	Short: "Exports all data from Conectordb into a new folder",
	Long: `Dumps the entire contents of ConnectorDB into a directory. This
allows you to upgrade ConnectorDB versions (by export old/import into new),
and to move ConnectorDB data between computers.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return ErrConfig
		}

		if len(args) < 2 {
			return errors.New("Must specify a non-existent directory into which to export")
		}
		if len(args) > 2 {
			return ErrTooManyArgs
		}

		cfg, err := config.LoadConfig(args[0])
		if err != nil {
			return err
		}

		setLogging(cfg)

		// Check if the export directory exists
		dir, err := filepath.Abs(args[1])
		if err != nil {
			return err
		}
		if util.PathExists(dir) {
			return errors.New("The given export location already exists")
		}

		// Open the ConnectorDB database
		db, err := connectordb.Open(cfg.Options())
		if err != nil {
			return err
		}
		defer db.Close()

		log.Info("Exporting To ", dir)

		if err = os.Mkdir(dir, 0700); err != nil {
			return err
		}

		usr, err := db.ReadAllUsers()
		if err != nil {
			return err
		}
		for u := range usr {
			log.Info("...Exporting ", usr[u].Name)
			usrdir := path.Join(dir, usr[u].Name)

			if err = os.Mkdir(usrdir, 0700); err != nil {
				return err
			}

			b, err := json.MarshalIndent(usr[u], "", "\t")
			if err != nil {
				return err
			}
			if err = ioutil.WriteFile(path.Join(usrdir, "user.json"), b, 0700); err != nil {
				return err
			}

			dev, err := db.ReadAllDevicesByUserID(usr[u].UserID)
			if err != nil {
				return err
			}
			for d := range dev {
				log.Info("............ ", usr[u].Name, "/", dev[d].Name)
				devdir := path.Join(usrdir, dev[d].Name)

				if err = os.Mkdir(devdir, 0700); err != nil {
					return err
				}

				b, err = json.MarshalIndent(dev[d], "", "\t")
				if err != nil {
					return err
				}
				if err = ioutil.WriteFile(path.Join(devdir, "device.json"), b, 0700); err != nil {
					return err
				}

				strm, err := db.ReadAllStreamsByDeviceID(dev[d].DeviceID)
				if err != nil {
					return err
				}
				for s := range strm {
					log.Debug("............ ", usr[u].Name, "/", dev[d].Name, "/", strm[s].Name)
					sdir := path.Join(devdir, strm[s].Name)

					if err = os.Mkdir(sdir, 0700); err != nil {
						return err
					}

					b, err = json.MarshalIndent(strm[s], "", "\t")
					if err != nil {
						return err
					}
					if err = ioutil.WriteFile(path.Join(sdir, "stream.json"), b, 0700); err != nil {
						return err
					}

					// Now we write the stream's data, and if it exists, the downlink stream
					dr, err := db.GetStreamIndexRangeByID(strm[s].StreamID, "", 0, 0, "")
					if err != nil {
						return err
					}

					if err = WriteStreamDataToFile(path.Join(sdir, "data.json"), dr); err != nil {
						return err
					}

					if strm[s].Downlink {
						// Now we write the stream's downlink
						dr, err := db.GetStreamIndexRangeByID(strm[s].StreamID, "downlink", 0, 0, "")
						if err != nil {
							return err
						}

						if err = WriteStreamDataToFile(path.Join(sdir, "downlink.json"), dr); err != nil {
							return err
						}
					}

				}
			}
		}

		// Everything is done. Now finally write the export struct to a file, so that import knows the
		// exporter version
		b, err := json.MarshalIndent(ExportInfo{
			Version:     2,
			ConnectorDB: connectordb.Version,
		}, "", "\t")
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(path.Join(dir, "connectordb.json"), b, 0700); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	RootCmd.AddCommand(ExportCmd)
}
