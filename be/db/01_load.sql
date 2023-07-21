INSERT INTO Entries (
	title, hash, posted, visible, entryId, version
)
VALUES
	('Lorem ipsum dolor sit amet',					'fd8bb7b97dc7f738f2e30700ea00b6b4', '2023-01-01 00:00:00', 1, 1, 1),
	('Praesent iaculis nisi',						'7efd22f6afd98fe6f289ca5c3bf57098', '2023-01-02 00:00:00', 1, 2, 1),
	('Sed sapien eros',								'f206dc881d6a062cd6fdf8bad0aadab3', '2023-01-03 00:00:00', 0, 3, 1),
	('Nulla placerat nunc placerat ex hendrerit',	'e56c56964b7b7edac33d3541d2309445', '2023-01-04 00:00:00', 1, 4, 1),
	('Cras eu faucibus felis',						'831e6a22e5fdfef393c42212ab2dc9fb', '2023-01-05 00:00:00', 1, 5, 1)
;

INSERT INTO Tags (
	tag, entryUid
)
VALUES 
	('abel', 1),
	('baker', 2),
	('charlie', 3),
	('dog', 4),
	('easy', 5),
	('foxtrot', 1),
	('foxtrot', 2),
	('george', 2),
	('george', 3),
	('how', 3),
	('how', 4),
	('item', 4),
	('item', 5),
	('jig', 1),
	('jig', 2),
	('jig', 3),
	('king', 3),
	('king', 4),
	('king', 5),
	('love', 1),
	('love', 2),
	('love', 3),
	('love', 4),
	('mike', 1),
	('mike', 2),
	('mike', 3),
	('mike', 4),
	('mike', 5)
;
