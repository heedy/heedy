package dashboard

import (
	"net/http"
	"os"
	"strconv"
	"testing"

	"github.com/heedy/heedy/api/golang/rest"
	"github.com/heedy/heedy/backend/assets"
	"github.com/heedy/heedy/backend/database"
	"github.com/heedy/heedy/backend/plugins/run"
	"github.com/jmoiron/sqlx/types"
	"github.com/stretchr/testify/require"
)

func testHandler() http.Handler {
	i := 0
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var j int
		err := rest.UnmarshalRequest(r, &j)
		evt := []DashboardEvent{
			{
				ObjectID: r.Header.Get("X-Heedy-Object"),
				Event:    "READ_ME",
			},
		}
		rest.WriteJSON(w, r, QueryResult{Events: &evt, Data: CompressedJSON(strconv.Itoa(i + j))}, err)
		i++
	})
}

type testHandlerGetter struct {
	m *run.Manager
}

func (t testHandlerGetter) GetHandler(uri string) (http.Handler, error) {
	return t.m.GetHandler("dashboard", uri)
}

func newAssets(t *testing.T) (*assets.Assets, func()) {
	a, err := assets.Open("", nil)
	require.NoError(t, err)
	os.RemoveAll("./test_db")
	a.FolderPath = "./test_db"
	sqla := "sqlite3://heedy.db?_journal=WAL&_fk=1"
	a.Config.SQL = &sqla

	assets.SetGlobal(a)
	return a, func() {
		os.RemoveAll("./test_db")
	}
}

func newDB(t *testing.T) (*database.AdminDB, func()) {
	a, cleanup := newAssets(t)

	err := database.Create(a)
	if err != nil {
		cleanup()
	}
	require.NoError(t, err)

	db, err := database.Open(a)
	require.NoError(t, err)

	// Add the hooks into core heedy code that are needed for dashboards to work.
	// dashboardtest is a plugin defined in the dashboard plugin's assets/heedy.conf
	run.Builtin.Add(&run.BuiltinRunner{
		Key:     "dashboardtest",
		Handler: testHandler(),
	})
	m := run.NewManager(db)
	runner := db.Assets().Config.Plugins["dashboardtest"].Run["test"]
	require.NoError(t, m.Start("dashboardtest", "test", &runner))

	// The dashboard global object needs to be initialized
	Dashboard, err = NewDashboardProcessor(db, db.Assets().Config.Plugins["dashboard"], testHandlerGetter{m})
	require.NoError(t, err)

	return db, cleanup
}

func newDBWithUser(t *testing.T) (*database.AdminDB, func()) {
	adb, cleanup := newDB(t)

	name := "test"
	passwd := "test"
	require.NoError(t, adb.CreateUser(&database.User{
		UserName: &name,
		Password: &passwd,
	}))
	return adb, cleanup
}

func newDBWithObjects(t *testing.T) (*database.AdminDB, string, string, func()) {
	db, cleanup := newDBWithUser(t)
	oname := "myobject"
	otype := "dashboard"
	uname := "test"
	oid1, err := db.CreateObject(&database.Object{
		Details: database.Details{
			Name: &oname,
		},
		Type:  &otype,
		Owner: &uname,
	})
	require.NoError(t, err)
	oid2, err := db.CreateObject(&database.Object{
		Details: database.Details{
			Name: &oname,
		},
		Type:  &otype,
		Owner: &uname,
	})
	require.NoError(t, err)
	return db, oid1, oid2, cleanup
}

func TestEvents(t *testing.T) {
	_, _, _, cleanup := newDBWithObjects(t)
	defer cleanup()

}

func TestCRUD(t *testing.T) {
	adb, oid1, oid2, cleanup := newDBWithObjects(t)
	defer cleanup()

	da, err := ReadDashboard(adb, "test", oid1, true)
	require.NoError(t, err)
	require.Len(t, da, 0)

	emptyObject := types.JSONText("{}")
	zeroObject := types.JSONText("0")
	oneObject := types.JSONText("1")

	require.Error(t, WriteDashboard(adb, "test", oid1, []DashboardElement{
		{
			Type:     "no_a_type",
			Query:    &zeroObject,
			Settings: &emptyObject,
		},
	}))

	require.Error(t, WriteDashboard(adb, "test", oid1, []DashboardElement{
		{
			Type:     "test",
			Query:    &emptyObject,
			Settings: &emptyObject,
		},
	}))

	err = WriteDashboard(adb, "test", oid1, []DashboardElement{
		{
			Type:     "test",
			Query:    &zeroObject,
			Settings: &emptyObject,
		},
	})
	require.NoError(t, err)

	da, err = ReadDashboard(adb, "test", oid2, true)
	require.NoError(t, err)
	require.Len(t, da, 0)
	da, err = ReadDashboard(adb, "test", oid1, true)
	require.NoError(t, err)
	require.Len(t, da, 1)
	b, err := da[0].Data.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "0", string(b))

	el, err := ReadDashboardElement(adb, "test", oid1, da[0].ID, false)
	require.NoError(t, err)
	require.Nil(t, el.Query)
	require.Nil(t, el.OnDemand)
	require.NotNil(t, el.Data)
	b, err = el.Data.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "0", string(b))

	err = WriteDashboard(adb, "test", oid1, []DashboardElement{
		{
			ID:    el.ID,
			Query: &oneObject,
		},
	})
	require.NoError(t, err)

	da, err = ReadDashboard(adb, "test", oid1, true)
	require.NoError(t, err)
	require.Len(t, da, 1)
	b, err = da[0].Data.MarshalJSON()
	require.NoError(t, err)
	require.Equal(t, "2", string(b))

	require.NoError(t, DeleteDashboardElement(adb, oid1, da[0].ID))

	_, err = ReadDashboardElement(adb, "test", oid1, da[0].ID, false)
	require.Error(t, err)

	da, err = ReadDashboard(adb, "test", oid1, true)
	require.NoError(t, err)
	require.Len(t, da, 0)

}
