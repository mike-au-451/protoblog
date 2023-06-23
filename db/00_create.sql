DROP TABLE IF EXISTS Entries;
CREATE TABLE Entries (
	id INTEGER PRIMARY KEY,
	title TEXT,
	body TEXT,
	posted TEXT,
	visible INTEGER DEFAULT 0,
	entryId INTEGER,
	version INTEGER,

	CONSTRAINT u_entry_version UNIQUE(entryId, version)
);
