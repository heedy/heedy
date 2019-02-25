package database

// This is the additions to the schema added when using sqlite.
// We use it mainly to implement search in the database
var sqliteAddonSchema = ``

/*

// This is the additions to the schema added when using sqlite.
// We use it mainly to implement search in the database
var sqliteAddonSchema = `

-- We use string ids, and fts requires int ids. This maps between the two
CREATE TABLE fts_idmap (
	ftsid INTEGER PRIMARY KEY,
	groupid VARCHAR UNIQUE
);
CREATE INDEX fts_groupid ON fts_idmap (groupid);


CREATE VIRTUAL TABLE group_search USING fts5(name,fullname,description,content='', tokenize = 'porter unicode61 remove_diacritics 1');

CREATE TRIGGER group_search_ai AFTER INSERT ON groups BEGIN
	INSERT INTO group_search(rowid, name,fullname,description) VALUES (new.id, new.name, new.fullname, new.description);
END;
CREATE TRIGGER group_search_ad AFTER DELETE ON groups BEGIN
	INSERT INTO group_search(group_search, rowid, name,fullname,description) VALUES('delete', old.id, old.name, old.fullname, old.description);
END;
CREATE TRIGGER group_search_au AFTER UPDATE ON groups BEGIN
	INSERT INTO group_search(group_search, rowid, name,fullname,description) VALUES('delete', old.id, old.name, old.fullname, old.description);
	INSERT INTO group_search(rowid, name,fullname,description) VALUES (new.id, new.name, new.fullname, new.description);
END;

`
*/
