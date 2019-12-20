package kv

import "github.com/heedy/heedy/backend/database"

type AdminUserKV struct {
	DB        *database.AdminDB
	ID        string
	Namespace string
}

func (k *AdminUserKV) Get() (map[string]interface{}, error) {
	return getKV(k.DB, `SELECT key,value FROM kv_user WHERE user=? AND namespace=?`, k.ID, k.Namespace)
}
func (k *AdminUserKV) Set(data map[string]interface{}) error {
	return setKV(k.DB, data, `DELETE FROM kv_user WHERE user=? AND namespace=?`, []interface{}{k.ID, k.Namespace}, `INSERT INTO kv_user(user,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminUserKV) Update(data map[string]interface{}) error {
	return updateKV(k.DB, data, `INSERT OR REPLACE INTO kv_user(user,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}

func (k *AdminUserKV) SetKey(key string, value interface{}) error {
	return setKey(k.DB, key, value, `INSERT OR REPLACE INTO kv_user(user,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminUserKV) GetKey(key string) (interface{}, error) {
	return getKey(k.DB, `SELECT value FROM kv_user WHERE user=? AND namespace=? AND key=?;`, k.ID, k.Namespace, key)
}
func (k *AdminUserKV) DelKey(key string) error {
	return runStatement(k.DB, "DELETE FROM kv_user WHERE user=? AND namespace=? AND key=?;", k.ID, k.Namespace, key)
}

type AdminAppKV struct {
	DB        *database.AdminDB
	ID        string
	Namespace string
}

func (k *AdminAppKV) Get() (map[string]interface{}, error) {
	return getKV(k.DB, `SELECT key,value FROM kv_app WHERE app=? AND namespace=?`, k.ID, k.Namespace)
}
func (k *AdminAppKV) Set(data map[string]interface{}) error {
	return setKV(k.DB, data, `DELETE FROM kv_app WHERE app=? AND namespace=?`, []interface{}{k.ID, k.Namespace}, `INSERT INTO kv_app(app,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminAppKV) Update(data map[string]interface{}) error {
	return updateKV(k.DB, data, `INSERT OR REPLACE INTO kv_app(app,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}

func (k *AdminAppKV) SetKey(key string, value interface{}) error {
	return setKey(k.DB, key, value, `INSERT OR REPLACE INTO kv_app(app,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminAppKV) GetKey(key string) (interface{}, error) {
	return getKey(k.DB, `SELECT value FROM kv_app WHERE app=? AND namespace=? AND key=?;`, k.ID, k.Namespace, key)
}
func (k *AdminAppKV) DelKey(key string) error {
	return runStatement(k.DB, "DELETE FROM kv_app WHERE app=? AND namespace=? AND key=?;", k.ID, k.Namespace, key)
}

type AdminObjectKV struct {
	DB        *database.AdminDB
	ID        string
	Namespace string
}

func (k *AdminObjectKV) Get() (map[string]interface{}, error) {
	return getKV(k.DB, `SELECT key,value FROM kv_object WHERE object=? AND namespace=?`, k.ID, k.Namespace)
}
func (k *AdminObjectKV) Set(data map[string]interface{}) error {
	return setKV(k.DB, data, `DELETE FROM kv_object WHERE object=? AND namespace=?`, []interface{}{k.ID, k.Namespace}, `INSERT INTO kv_object(object,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminObjectKV) Update(data map[string]interface{}) error {
	return updateKV(k.DB, data, `INSERT OR REPLACE INTO kv_object(object,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}

func (k *AdminObjectKV) SetKey(key string, value interface{}) error {
	return setKey(k.DB, key, value, `INSERT OR REPLACE INTO kv_object(object,namespace,key,value) VALUES (?,?,?,?)`, k.ID, k.Namespace)
}
func (k *AdminObjectKV) GetKey(key string) (interface{}, error) {
	return getKey(k.DB, `SELECT value FROM kv_object WHERE object=? AND namespace=? AND key=?;`, k.ID, k.Namespace, key)
}
func (k *AdminObjectKV) DelKey(key string) error {
	return runStatement(k.DB, "DELETE FROM kv_object WHERE object=? AND namespace=? AND key=?;", k.ID, k.Namespace, key)
}
