package timeseries

import (
	"errors"

	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/heedy/pipescript/datasets/interpolators"
	"github.com/heedy/pipescript/transforms"
	"github.com/klauspost/compress/zstd"
	"github.com/mitchellh/mapstructure"
	"github.com/sirupsen/logrus"
)

const PluginName = "timeseries"

// The global timeseries DB object that is initialized on database start
var TSDB TimeseriesDB

// SQLUpdater is in the format expected by Heedy to update the database
func SQLUpdater(db *database.AdminDB, i *run.Info, h run.BuiltinHelper, curversion int) error {
	if curversion == SQLVersion {
		return nil
	}
	if curversion != 0 {
		return errors.New("Timeseries database version incompatible")
	}
	_, err := db.ExecUncached(sqlSchema)
	return err
}

// StartTimeseries prepares the plugin by initializing the database
func StartTimeseries(db *database.AdminDB, i *run.Info, h run.BuiltinHelper) error {
	err := run.WithVersion(PluginName, SQLVersion, SQLUpdater)(db, i, h)
	if err != nil {
		return err
	}

	tsc, ok := db.Assets().Config.Plugins["timeseries"]
	if !ok {
		return errors.New("Could not find timeseries plugin configuration")
	}

	err = mapstructure.Decode(tsc.Settings, &TSDB)
	if err != nil {
		return err
	}
	TSDB.DB = db

	if TSDB.BatchSize <= 1 || TSDB.MaxBatchSize <= TSDB.BatchSize {
		return errors.New("Timeseries batch size must be at least 1, and max batch size must be more than batch size")
	}

	if TSDB.BatchCompressionLevel < 0 {
		logrus.WithField("plugin", "timeseries").Warn("Batch compression turned off, use this only on test databases!")
		zencoder = nil // set to nil means no compression
	} else if TSDB.BatchCompressionLevel > 3 {
		return errors.New("Timeseries currently doesn't support compression rates > 3")
	} else {
		zencoder, err = zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.EncoderLevel(TSDB.BatchCompressionLevel)))
	}

	return err
}

// This is not needed for normal plugins. The init simply registers the plugin with heedy internals
// for when it is compiled directly into the main heedy executable.
func init() {
	// Register all pipescript transforms
	transforms.Register()
	interpolators.Register()

	// Initialize the plugin
	run.Builtin.Add(&run.BuiltinRunner{
		Key:     PluginName,
		Start:   StartTimeseries,
		Handler: Handler,
	})
	// Runs schema creation on database create instead of on first start
	database.AddCreateHook(run.WithNilInfo(run.WithVersion(PluginName, SQLVersion, SQLUpdater)))
}
