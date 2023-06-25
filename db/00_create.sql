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

DROP TABLE IF EXISTS Tags;
CREATE TABLE Tags (
	id INTEGER PRIMARY KEY,
	tag TEXT,
	entryUid INTEGER REFERENCES Entries(id),

	-- should be ok considering there will be very few inserts.
	CONSTRAINT u_tag_entry UNIQUE(tag, entryUid)
);

